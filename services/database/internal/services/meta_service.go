package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type MetaService struct {
	db  *pgx.Conn
	log *zap.Logger
}

func NewMetaService(db *pgx.Conn, log *zap.Logger) *MetaService {
	return &MetaService{
		db:  db,
		log: log,
	}
}

type TableMeta struct {
	Name     string `json:"name"`
	Schema   string `json:"schema"`
	RowCount int64  `json:"row_count"`
	Size     string `json:"size"`
	HasRLS   bool   `json:"has_rls"`
}

type CreateTableRequest struct {
	Name    string         `json:"name"`
	Schema  string         `json:"schema"`
	Columns []ColumnDefinition `json:"columns"`
}

type ColumnDefinition struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsNullable bool   `json:"is_nullable"`
	IsPrimary  bool   `json:"is_primary"`
	Default    string `json:"default,omitempty"`
}

type FunctionMeta struct {
	Name       string `json:"name"`
	Schema     string `json:"schema"`
	ReturnType string `json:"return_type"`
	Language   string `json:"language"`
	Definition string `json:"definition"`
}

type PolicyMeta struct {
	Name       string   `json:"name"`
	Schema     string   `json:"schema"`
	Table      string   `json:"table"`
	Action     string   `json:"action"` // ALL, SELECT, INSERT, etc.
	Roles      []string `json:"roles"`
	Qualifier  string   `json:"qualifier"` // USING / WITH CHECK
}

// GetTables returns a list of tables and their metadata in the specified schema
func (s *MetaService) GetTables(ctx context.Context, schema string) ([]TableMeta, error) {
	query := `
		SELECT 
			c.relname as name,
			n.nspname as schema,
			COALESCE(s.n_live_tup, 0) as row_count,
			pg_size_pretty(pg_total_relation_size(c.oid)) as size,
			c.relrowsecurity as has_rls
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_stat_user_tables s ON s.relid = c.oid
		WHERE n.nspname = $1 AND c.relkind = 'r'
		ORDER BY c.relname;
	`
	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableMeta
	for rows.Next() {
		var t TableMeta
		if err := rows.Scan(&t.Name, &t.Schema, &t.RowCount, &t.Size, &t.HasRLS); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// RunQuery executes an arbitrary SQL string and returns the raw rows
// In OmniBase, this requires Service Role token (Admin only).
func (s *MetaService) RunQuery(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	
	fields := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		
		rowMap := make(map[string]interface{})
		for i, field := range fields {
			rowMap[string(field.Name)] = values[i]
		}
		results = append(results, rowMap)
	}
	
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return results, nil
}

// ResolveGraphQL calls the pg_graphql extension to resolve a query
func (s *MetaService) ResolveGraphQL(ctx context.Context, query string, variables map[string]interface{}, userID, role string) (interface{}, error) {
	var resultBytes []byte
	
	varsStr := "{}"
	if variables != nil {
		varsBytes, _ := json.Marshal(variables)
		varsStr = string(varsBytes)
	}

	// Use a transaction to set the role and claims for RLS
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		// Set the role
		if role != "" {
			if _, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL ROLE %s", pgx.Identifier{role}.Sanitize())); err != nil {
				return err
			}
		}

		// Set the claims (PostgREST style)
		if userID != "" || role != "" {
			claims := map[string]interface{}{
				"sub":  userID,
				"role": role,
			}
			claimsJSON, _ := json.Marshal(claims)
			if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", string(claimsJSON)); err != nil {
				return err
			}
		}

		return tx.QueryRow(ctx, "SELECT graphql.resolve(query := $1, variables := $2)", query, varsStr).Scan(&resultBytes)
	})

	if err != nil {
		return nil, err
	}

	var parsed interface{}
	if err := json.Unmarshal(resultBytes, &parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

// CreateTable creates a new table in the specified schema
func (s *MetaService) CreateTable(ctx context.Context, req CreateTableRequest) error {
	if req.Schema == "" {
		req.Schema = "public"
	}

	sql := fmt.Sprintf("CREATE TABLE %s.%s (", req.Schema, req.Name)
	var primaryKeys []string
	var columns []string

	for _, col := range req.Columns {
		colDef := fmt.Sprintf("%s %s", col.Name, col.Type)
		if !col.IsNullable {
			colDef += " NOT NULL"
		}
		if col.Default != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		columns = append(columns, colDef)
		if col.IsPrimary {
			primaryKeys = append(primaryKeys, col.Name)
		}
	}

	sql += strings.Join(columns, ", ")
	if len(primaryKeys) > 0 {
		sql += fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
	}
	sql += ");"

	_, resErr := s.db.Exec(ctx, sql)
	if resErr == nil {
		s.ReloadSchemaCache(ctx)
	}
	return resErr
}

// CreateProject creates a new database schema for a project
func (s *MetaService) CreateProject(ctx context.Context, name string) error {
	// 1. Create schema
	_, err := s.db.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", name))
	if err != nil {
		return err
	}

	// 2. Grant usage to public/anon (Phase 1 simplicity)
	// In production, we'd create dedicated roles per project
	_, err = s.db.Exec(ctx, fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO authenticated, anon", name))
	if err != nil {
		return err
	}

	_, resErr := s.db.Exec(ctx, fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON TABLES TO authenticated, anon", name))
	if resErr == nil {
		s.ReloadSchemaCache(ctx)
	}
	return resErr
}

// SetRLSEnabled enabled or disables RLS on a table
func (s *MetaService) SetRLSEnabled(ctx context.Context, schema, table string, enabled bool) error {
	action := "ENABLE"
	if !enabled {
		action = "DISABLE"
	}
	_, err := s.db.Exec(ctx, fmt.Sprintf("ALTER TABLE %s.%s %s ROW LEVEL SECURITY", schema, table, action))
	return err
}

func (s *MetaService) GetFunctions(ctx context.Context, schema string) ([]FunctionMeta, error) {
	query := `
		SELECT 
			p.proname as name,
			n.nspname as schema,
			pg_get_function_result(p.oid) as return_type,
			l.lanname as language,
			pg_get_functiondef(p.oid) as definition
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		JOIN pg_language l ON l.oid = p.prolang
		WHERE n.nspname = $1
		ORDER BY p.proname;
	`
	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fns []FunctionMeta
	for rows.Next() {
		var f FunctionMeta
		if err := rows.Scan(&f.Name, &f.Schema, &f.ReturnType, &f.Language, &f.Definition); err != nil {
			return nil, err
		}
		fns = append(fns, f)
	}
	return fns, nil
}

func (s *MetaService) GetPolicies(ctx context.Context, schema string) ([]PolicyMeta, error) {
	query := `
		SELECT 
			policyname as name,
			schemaname as schema,
			tablename as table,
			cmd as action,
			roles::text[] as roles,
			qual as qualifier
		FROM pg_policies
		WHERE schemaname = $1
		ORDER BY tablename, policyname;
	`
	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []PolicyMeta
	for rows.Next() {
		var p PolicyMeta
		if err := rows.Scan(&p.Name, &p.Schema, &p.Table, &p.Action, &p.Roles, &p.Qualifier); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func (s *MetaService) GetSchemas(ctx context.Context) ([]string, error) {
	query := `
		SELECT nspname 
		FROM pg_namespace 
		WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'storage', 'auth', 'realtime', 'omnibase')
		AND nspname NOT LIKE 'pg_toast%'
		ORDER BY nspname;
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		schemas = append(schemas, name)
	}
	return schemas, nil
}

// ReloadSchemaCache notifies PostgREST to refresh its schema cache
func (s *MetaService) ReloadSchemaCache(ctx context.Context) error {
	_, err := s.db.Exec(ctx, "NOTIFY pgrst, 'reload schema'")
	return err
}



