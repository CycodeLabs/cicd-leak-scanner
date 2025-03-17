package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/CycodeLabs/cicd-leak-scanner/pkg/config"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/decoder"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/github"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/output"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel string
	orgName  string
	repoName string
	token    string
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

			if token == "" {
				return fmt.Errorf("GitHub token is required")
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

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("Error executing command: %v", err)
	}

	return nil
}

func run() error {
	ctx := context.Background()
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
			log.Error().Msgf("Error searching workflows: %v", err)
			return fmt.Errorf("Error searching workflows: %v", err)
		}

		for result := range codeResults {
			owner := result.GetRepository().GetOwner().GetLogin()
			repo := result.GetRepository().GetName()
			workflowFile := strings.TrimPrefix(result.GetPath(), ".github/workflows/")

			log.Info().Str("Org", owner).Str("Repo", repo).Str("Workflow", workflowFile).Msg("Processing workflow")

			run, err := githubClient.GetLatestSuccessfulWorkflowRun(ctx, owner, repo, workflowFile)
			if err != nil || run == nil {
				log.Warn().Msgf("No successful runs found for %s/%s", owner, repo)
				continue
			}

			log.Debug().Str("Workflow", workflowFile).Int64("Run", run.GetID()).Msg("Processing run")

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

			log.Debug().Int("logContent", len(logContent)).Msg("Fetched log content")

			pattern := regexp.MustCompile(rule.Regex)
			matches := pattern.FindStringSubmatch(logContent)
			if len(matches) == 0 {
				continue
			}

			for _, dec := range rule.Decoders {
				decoder, err := decoder.New(dec.Id)
				if err != nil {
					log.Warn().Msgf("Error creating decoder: %v", err)
					continue
				}

				decoded, err := decoder.Decode(matches[1], dec.Repeat)
				if err != nil {
					log.Warn().Msgf("Error decoding secret: %v", err)
					continue
				}

				log.Info().Msg("Found secret")
				if err := outputClient.Write(owner, repo, workflowFile, decoded); err != nil {
					log.Warn().Msgf("Error writing secret: %v", err)
				}
			}
		}
	}

	return nil
}
