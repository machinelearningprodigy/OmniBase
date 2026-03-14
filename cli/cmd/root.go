package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

var (
	url    string
	apiKey string
)

var rootCmd = &cobra.Command{
	Use:   "omnibase",
	Short: "OmniBase CLI — manage your OmniBase project",
	Long:  `OmniBase is an open-source Backend-as-a-Service. Use this CLI to scaffold projects, generate types, run migrations, and more.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&url, "url", getEnv("OMNIBASE_URL", "http://localhost:8000"), "OmniBase API URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "key", getEnv("OMNIBASE_ANON_KEY", ""), "OmniBase anon key (or service role for admin)")
}
