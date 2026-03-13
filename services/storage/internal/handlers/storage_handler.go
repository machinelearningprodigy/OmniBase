package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"github.com/machinelearningprodigy/OmniBase/storage/internal/services"
	"go.uber.org/zap"
)

type StorageHandler struct {
	service *services.StorageService
	jwt     *jwt.Manager
	log     *zap.Logger
}

func NewStorageHandler(service *services.StorageService, jwtManager *jwt.Manager, log *zap.Logger) *StorageHandler {
	return &StorageHandler{
		service: service,
		jwt:     jwtManager,
		log:     log,
	}
}

// Auth Middleware purely for the storage handlers
func (h *StorageHandler) RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid authorization token"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.jwt.Verify(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	c.Locals("user_id", claims.UserID)
	return c.Next()
}

func (h *StorageHandler) ListBuckets(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	buckets, err := h.service.ListBuckets(c.Context(), userID)
	if err != nil {
		h.log.Error("list buckets error", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to list buckets"})
	}
	return c.JSON(buckets)
}

func (h *StorageHandler) CreateBucket(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var body struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Public bool   `json:"public"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if body.Name == "" {
		body.Name = body.ID
	}

	if err := h.service.CreateBucket(c.Context(), userID, body.ID, body.Name, body.Public); err != nil {
		h.log.Error("create bucket error", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to create bucket"})
	}
	return c.JSON(fiber.Map{"message": "bucket created", "name": body.Name})
}

func (h *StorageHandler) DeleteBucket(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	bucketID := c.Params("id")

	if err := h.service.DeleteBucket(c.Context(), userID, bucketID); err != nil {
		h.log.Error("delete bucket error", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete bucket"})
	}
	return c.JSON(fiber.Map{"message": "bucket deleted"})
}

func (h *StorageHandler) UploadObject(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	bucketID := c.Params("bucket")
	path := c.Params("*")

	contentType := c.Get("Content-Type", "application/octet-stream")
	
	// Read full body in Phase 1 (Memory limit applies in Fiber)
	body := c.Body()
	size := int64(len(body))

	reader := strings.NewReader(string(body))

	if err := h.service.UploadObject(c.Context(), userID, bucketID, path, reader, size, contentType); err != nil {
		h.log.Error("upload error", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to upload object"})
	}
	return c.JSON(fiber.Map{"Key": fmt.Sprintf("%s/%s", bucketID, path)})
}

func (h *StorageHandler) DeleteObjects(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	bucketID := c.Params("bucket")

	var body struct {
		Prefixes []string `json:"prefixes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := h.service.DeleteObjects(c.Context(), userID, bucketID, body.Prefixes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete objects"})
	}
	return c.JSON(fiber.Map{"message": "objects deleted"})
}

func (h *StorageHandler) GetPublicObject(c *fiber.Ctx) error {
	bucketID := c.Params("bucket")
	path := c.Params("*")

	reader, size, contentType, err := h.service.GetPublicObject(c.Context(), bucketID, path)
	if err != nil {
		return c.Status(404).SendString("object not found or bucket is private")
	}
	defer reader.Close()

	c.Set("Content-Type", contentType)
	c.Set("Content-Length", fmt.Sprintf("%d", size))
	c.Set("Cache-Control", "public, max-age=3600")

	// Fiber's SendStream takes an io.Reader
	return c.SendStream(reader)
}

func (h *StorageHandler) CreateSignedURL(c *fiber.Ctx) error {
	bucketID := c.Params("bucket")
	path := c.Params("*")

	var body struct {
		ExpiresIn int `json:"expiresIn"`
	}
	if err := c.BodyParser(&body); err != nil {
		body.ExpiresIn = 3600
	}

	if body.ExpiresIn <= 0 {
		body.ExpiresIn = 3600
	}

	_, err := h.service.CreateSignedURL(c.Context(), bucketID, path, body.ExpiresIn)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := h.jwt.IssueStorageToken(bucketID, path, body.ExpiresIn)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	scheme := c.Protocol()
	host := c.Hostname()
	fullURL := fmt.Sprintf("%s://%s/storage/v1/object/sign/%s/%s?token=%s", scheme, host, bucketID, path, token)

	return c.JSON(fiber.Map{"signedURL": fullURL})
}

func (h *StorageHandler) GetSignedObject(c *fiber.Ctx) error {
	bucketID := c.Params("bucket")
	path := c.Params("*")
	tokenStr := c.Query("token")

	if tokenStr == "" {
		return c.Status(401).SendString("missing token")
	}

	claims, err := h.jwt.VerifyStorageToken(tokenStr)
	if err != nil {
		return c.Status(403).SendString("invalid or expired token")
	}

	expectedSub := fmt.Sprintf("storage:%s:%s", bucketID, path)
	if claims.Subject != expectedSub {
		return c.Status(403).SendString("token does not match requested object")
	}

	reader, size, contentType, err := h.service.GetObjectDirect(c.Context(), bucketID, path)
	if err != nil {
		return c.Status(404).SendString("object not found")
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Length", fmt.Sprintf("%d", size))
	c.Set("Cache-Control", "public, max-age=3600")

	return c.SendStream(reader)
}
