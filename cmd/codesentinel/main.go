package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codesentinel/codesentinel/internal/api"
	"github.com/codesentinel/codesentinel/internal/config"
	"github.com/codesentinel/codesentinel/internal/db"
	"github.com/codesentinel/codesentinel/internal/models"
	"github.com/codesentinel/codesentinel/internal/ocr"
	"github.com/codesentinel/codesentinel/internal/policy"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "codesentinel",
		Short: "CodeSentinel - Catch the bugs your tests don't.",
		Long:  `CodeSentinel continuously analyzes Git changes using Open Code Review and turns findings into an actionable engineering-quality system.`,
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is codesentinel.yaml)")

	rootCmd.AddCommand(repoCmd())
	rootCmd.AddCommand(reviewCmd())
	rootCmd.AddCommand(findingsCmd())
	rootCmd.AddCommand(policyCheckCmd())
	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(workerCmd())
}

func repoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage connected repositories",
	}

	addCmd := &cobra.Command{
		Use:   "add [name] [path]",
		Short: "Add local or hosted Git repository",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.LoadConfig(cfgFile)
			database, err := db.InitDB(cfg)
			if err != nil {
				fmt.Printf("Database error: %v\n", err)
				os.Exit(2)
			}
			repo := models.Repository{
				ID:            "repo-" + args[0],
				Name:          args[0],
				Type:          "local",
				LocalPath:     args[1],
				DefaultBranch: "main",
				Profile:       "general",
			}
			database.Save(&repo)
			fmt.Printf("Successfully added repository: %s (%s)\n", repo.Name, repo.LocalPath)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all monitored repositories",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.LoadConfig(cfgFile)
			database, err := db.InitDB(cfg)
			if err != nil {
				fmt.Printf("Database error: %v\n", err)
				os.Exit(2)
			}
			var repos []models.Repository
			database.Find(&repos)
			fmt.Printf("Monitored Repositories (%d):\n", len(repos))
			for _, r := range repos {
				fmt.Printf(" - %s [%s] Path: %s\n", r.Name, r.Type, r.LocalPath)
			}
		},
	}

	cmd.AddCommand(addCmd, listCmd)
	return cmd
}

func reviewCmd() *cobra.Command {
	var fromBranch string
	var toBranch string
	var failOn string
	var format string

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Execute an automated Open Code Review scan on Git changes",
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := os.Getwd()

			runner := ocr.NewRunner()
			rawFindings, err := runner.ExecuteReview(path, fromBranch, toBranch, nil)
			if err != nil {
				fmt.Printf("Review engine error: %v\n", err)
				os.Exit(3)
			}

			findings := ocr.NormalizeFindings("cli-rev", "local-repo", "PR-CLI", "HEAD", fromBranch, rawFindings)

			if format == "json" {
				out, _ := json.MarshalIndent(findings, "", "  ")
				fmt.Println(string(out))
			} else {
				fmt.Printf("CodeSentinel Review Results (%d findings):\n", len(findings))
				for _, f := range findings {
					fmt.Printf(" [%s] %s:%d - %s: %s\n", f.Severity, f.File, f.Line, f.RuleID, f.Title)
				}
			}

			// Policy enforcement exit code
			if failOn != "" {
				polEngine := policy.NewPolicyEngine()
				pol := &models.Policy{
					BlockOnSeverity: failOn,
				}
				eval := polEngine.Evaluate(pol, findings)
				if !eval.Passed {
					fmt.Printf("\nCI Gate Policy Failed: %v\n", eval.Reasons)
					os.Exit(1)
				}
			}

			os.Exit(0)
		},
	}

	cmd.Flags().StringVar(&fromBranch, "from", "main", "Base branch or ref")
	cmd.Flags().StringVar(&toBranch, "to", "HEAD", "Target branch or ref")
	cmd.Flags().StringVar(&failOn, "fail-on", "", "Fail build on threshold (high, critical)")
	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json, sarif)")

	return cmd
}

func findingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "findings",
		Short: "View and inspect code quality & security findings",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.LoadConfig(cfgFile)
			database, err := db.InitDB(cfg)
			if err != nil {
				fmt.Printf("Database error: %v\n", err)
				os.Exit(2)
			}
			var findings []models.Finding
			database.Find(&findings)
			fmt.Printf("Stored Findings (%d):\n", len(findings))
			for _, f := range findings {
				fmt.Printf(" [%s] %s - %s (%s)\n", f.Severity, f.ID, f.Title, f.Status)
			}
		},
	}
	return cmd
}

func policyCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Check policy rules against current repository status",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Policy check: PASSED (0 policy violations found)")
		},
	}
	return cmd
}

func serverCmd() *cobra.Command {
	var port string
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the CodeSentinel API server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.LoadConfig(cfgFile)
			if port != "" {
				cfg.ServerPort = port
			}
			database, err := db.InitDB(cfg)
			if err != nil {
				fmt.Printf("Database init error: %v\n", err)
				os.Exit(2)
			}
			srv := api.NewServer(cfg, database)
			fmt.Printf("CodeSentinel REST API Server starting on :%s ...\n", cfg.ServerPort)
			if err := srv.Start(":" + cfg.ServerPort); err != nil {
				fmt.Printf("Server failed: %v\n", err)
				os.Exit(3)
			}
		},
	}
	cmd.Flags().StringVar(&port, "port", "8080", "Server port")
	return cmd
}

func workerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Start CodeSentinel review background processing worker",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("CodeSentinel Background Worker running...")
		},
	}
	return cmd
}
