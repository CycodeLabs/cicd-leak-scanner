package github

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const (
	maxRedirects = 5
)

type GitHub struct {
	ctx    context.Context
	client *github.Client
}

func New(ctx context.Context, token string) *GitHub {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return &GitHub{
		ctx:    ctx,
		client: github.NewClient(tc),
	}
}

func (g *GitHub) SearchWorkflows(query string) (<-chan *github.CodeResult, error) {
	ch := make(chan *github.CodeResult)

	go func() {
		defer close(ch)

		opts := &github.SearchOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		for {
			results, resp, err := g.client.Search.Code(g.ctx, query, opts)
			if err != nil {
				if rlErr, ok := err.(*github.RateLimitError); ok {
					sleepDuration := time.Until(rlErr.Rate.Reset.Time)
					if sleepDuration <= 0 {
						sleepDuration = 1 * time.Second
					}

					log.Warn().Msgf("Rate limited, waiting for %v", sleepDuration)
					time.Sleep(sleepDuration)
					continue
				}

				log.Error().Msgf("Error fetching search results: %v", err)
				return
			}

			for _, result := range results.CodeResults {
				select {
				case <-g.ctx.Done():
					return
				case ch <- result:
				}
			}

			if resp.NextPage == 0 {
				log.Info().Int("page", opts.Page).Int("Results", len(results.CodeResults)).Msg("Fetching search results")
				break
			}

			if resp.StatusCode == http.StatusOK {
				log.Info().Int("page", opts.Page).Int("Results", len(results.CodeResults)).Msg("Fetching search results")
				opts.Page = resp.NextPage
				continue
			}
		}
	}()

	return ch, nil
}

func (g *GitHub) GetLatestSuccessfulWorkflowRuns(ctx context.Context, owner, repo, workflowFileName string, numRuns int) ([]*github.WorkflowRun, error) {
	var allRuns []*github.WorkflowRun
	perPage := 100
	pages := (numRuns + perPage - 1) / perPage

	for page := 1; page <= pages; page++ {
		log.Trace().Int("Page", page).Msg("Scanning workflow runs")

		remaining := numRuns - len(allRuns)
		if remaining < perPage {
			perPage = remaining
		}

		runs, _, err := g.client.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, workflowFileName, &github.ListWorkflowRunsOptions{
			Status: "success",
			ListOptions: github.ListOptions{
				PerPage: perPage,
				Page:    page,
			},
		})

		if err != nil {
			return nil, err
		}

		allRuns = append(allRuns, runs.WorkflowRuns...)

		if len(runs.WorkflowRuns) < perPage {
			break
		}
	}

	return allRuns, nil
}

func (g *GitHub) GetJobLogs(owner, repo string, runID int64) (string, error) {
	url, _, err := g.client.Actions.GetWorkflowRunLogs(g.ctx, owner, repo, runID, maxRedirects)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (g *GitHub) GetJobLogsContent(logURL string) (string, error) {
	resp, err := http.Get(logURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return "", err
	}

	var logs strings.Builder
	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			continue
		}
		io.Copy(&logs, rc)
		rc.Close()
	}

	return logs.String(), nil
}
