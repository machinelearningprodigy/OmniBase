package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a new OmniBase project",
	Long:  `Creates .omnibase directory, example .env, and optional migrations folder.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}
	base := filepath.Join(dir, ".omnibase")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return fmt.Errorf("create .omnibase: %w", err)
	}
	migrationsDir := filepath.Join(base, "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		return fmt.Errorf("create migrations dir: %w", err)
	}

	envPath := filepath.Join(dir, ".env.example")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		content := `# OmniBase local development
OMNIBASE_URL=http://localhost:8000
OMNIBASE_ANON_KEY=your-anon-key-from-dashboard-settings
`
		if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write .env.example: %w", err)
		}
		fmt.Println("Created .env.example")
	}

	readmePath := filepath.Join(base, "README.md")
	content := `# OmniBase project

- Get your anon key from the dashboard: http://localhost:3001/settings (after sign-in)
- Run migrations: omnibase db migrate
- Generate types: omnibase gen types
`
	if err := os.WriteFile(readmePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write .omnibase/README.md: %w", err)
	}

	fmt.Printf("Initialized OmniBase project in %s\n", dir)
	fmt.Println("  .omnibase/")
	fmt.Println("  .omnibase/migrations/")
	fmt.Println("  .omnibase/README.md")
	return nil
}
