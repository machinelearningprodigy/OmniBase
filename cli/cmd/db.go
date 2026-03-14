package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var dbMigratePath string

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run pending SQL migrations",
	Long:  `Runs SQL files from .omnibase/migrations or --path in order. Requires service role key.`,
	RunE:  runDbMigrate,
}

func init() {
	dbCmd := &cobra.Command{Use: "db", Short: "Database commands"}
	dbCmd.AddCommand(dbMigrateCmd)
	rootCmd.AddCommand(dbCmd)
	dbMigrateCmd.Flags().StringVar(&dbMigratePath, "path", ".omnibase/migrations", "Migrations directory")
}

func runDbMigrate(cmd *cobra.Command, args []string) error {
	if apiKey == "" {
		return fmt.Errorf("service role key required for migrations: set OMNIBASE_ANON_KEY or --key")
	}
	entries, err := os.ReadDir(dbMigratePath)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	for _, name := range files {
		path := filepath.Join(dbMigratePath, name)
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if err := runQuery(string(sql), name); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		fmt.Printf("  ✓ %s\n", name)
	}
	if len(files) == 0 {
		fmt.Println("No migration files found in", dbMigratePath)
		return nil
	}
	fmt.Printf("Ran %d migration(s)\n", len(files))
	return nil
}

func runQuery(sql, migrationName string) error {
	body := map[string]interface{}{"query": sql, "migration_name": migrationName}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", url+"/pg/query", strings.NewReader(string(jsonBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
