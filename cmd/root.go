package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/CycodeLabs/cicd-leak-scanner/pkg/config"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/decoder"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/github"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/output"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/output/stdout"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel string
	orgName  string
	repoName string
	token    string

	orgScanned      = make(map[string]bool)
	repoScanned     = make(map[string]bool)
	workflowScanned = make(map[string]bool)
	secretsFound    = make(map[string]bool)
	runsScanned     = make(map[int64]bool)
)

func Execute() error {
	var rootCmd = &cobra.Command{
		Use:   "cicd-leak-scanner",
		Short: "CI/CD Leak Scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch logLevel {
			case "info":
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			case "debug":
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			case "trace":
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			default:
				log.Error().Msgf("Invalid log level: %s", logLevel)
			}

			if err := run(); err != nil {
				return fmt.Errorf("Error running crawler: %v", err)
			}

			return nil
		},
	}

	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log Level (info, debug, trace)")
	rootCmd.Flags().StringVarP(&token, "github-token", "t", "", "GitHub token")
	rootCmd.Flags().StringVarP(&orgName, "org-name", "o", "", "GitHub organization")
	rootCmd.Flags().StringVarP(&repoName, "repo-name", "r", "", "GitHub repository (e.g. example-org/example-repo)")

	rootCmd.MarkFlagRequired("github-token")

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("Error executing command: %v", err)
	}

	return nil
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info().Msg("Received interrupt signal, shutting down gracefully")
		cancel()
	}()

	githubClient := github.New(ctx, token)
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("Error loading config: %v", err)
	}

	outputClient, err := output.New(cfg.Output.Method, output.WithFilename(cfg.Output.Filename))
	if err != nil {
		return fmt.Errorf("Error creating outputClient: %v", err)
	}

	for _, rule := range cfg.Rules {
		log.Info().Str("Rule", rule.Name).Msg("Processing rule")

		query := rule.Query
		if orgName != "" {
			query = fmt.Sprintf("%s org:%s", query, orgName)
		}

		if repoName != "" {
			query = fmt.Sprintf("%s repo:%s", query, repoName)
		}

		codeResults, err := githubClient.SearchWorkflows(query)
		if err != nil {
			return fmt.Errorf("Error searching workflows: %v", err)
		}

		for result := range codeResults {
			owner := result.GetRepository().GetOwner().GetLogin()
			orgScanned[owner] = true

			repo := result.GetRepository().GetName()
			repoScanned[fmt.Sprintf("%s/%s", owner, repo)] = true

			workflowFile := strings.TrimPrefix(result.GetPath(), ".github/workflows/")
			workflowScanned[workflowFile] = true

			log.Info().Str("Org", owner).Str("Repo", repo).Str("Workflow", workflowFile).Msg("Found workflow which matched the rule conditions")

			runs, err := githubClient.GetLatestSuccessfulWorkflowRuns(ctx, owner, repo, workflowFile, cfg.Scanner.WorkflowRunsToScan)
			if err != nil || len(runs) == 0 {
				log.Warn().Msgf("No successful runs found for %s/%s", owner, repo)
				continue
			}

			log.Info().Msgf("Scanning last %d successful runs", len(runs))

			for _, run := range runs {
				log.Debug().Str("Workflow", workflowFile).Int64("RunID", run.GetID()).Msg("Processing run")
				runsScanned[run.GetID()] = true

				logURL, err := githubClient.GetJobLogs(owner, repo, run.GetID())
				if err != nil {
					log.Warn().Msgf("Error fetching log URL: %v", err)
					continue
				}
				log.Debug().Msgf("Fetching logs from %s", logURL)

				logContent, err := githubClient.GetJobLogsContent(logURL)
				if err != nil {
					log.Printf("Error fetching log content: %v", err)
					continue
				}

				log.Info().Int("LogContentLen", len(logContent)).Msgf("Fetched log content for run %d", run.GetID())

				pattern := regexp.MustCompile(rule.Regex)
				matches := pattern.FindStringSubmatch(logContent)
				if len(matches) == 0 {
					continue
				}

				secret := matches[1]
				for _, dec := range rule.Decoders {
					decoder, err := decoder.New(dec.Id)
					if err != nil {
						log.Warn().Msgf("Error creating decoder: %v", err)
						secret = ""
						break
					}

					secret, err = decoder.Decode(secret, dec.Repeat)
					if err != nil {
						log.Warn().Msgf("Error decoding secret: %v", err)
						break
					}
				}

				if secret == "" {
					continue
				}

				secretsFound[secret] = true
				log.Info().Msg("Found secret in build logs")

				if err := outputClient.Write(owner, repo, workflowFile, run.GetID(), secret); err != nil {
					log.Warn().Msgf("Error writing secret: %v", err)
				}
			}
		}
	}

	stdout.PrintSummary(orgScanned, repoScanned, workflowScanned, runsScanned, len(secretsFound))

	return nil
}
