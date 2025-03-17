package stdout

import "fmt"

type StdOutput struct{}

func New() (*StdOutput, error) {
	return &StdOutput{}, nil
}

func (s *StdOutput) Write(org string, repo string, workflow string, run int64, secret string) error {
	output := fmt.Sprintf("=========== %s/%s ===========\n", org, repo)
	output += fmt.Sprintf("Workflow: %s\n", workflow)
	output += fmt.Sprintf("Workflow Run: %d\n", run)
	output += fmt.Sprintf("Secret: %s\n", secret)
	output += "===============================\n"

	fmt.Println(output)
	return nil
}

func PrintSummary(orgScanned map[string]bool, repoScanned map[string]bool, workflowScanned map[string]bool, runsScanned map[int64]bool, secretsFound int) {
	out := fmt.Sprintf("\n=========== Summary ===========\n")
	out += fmt.Sprintf("Organizations Scanned: %d\n", len(orgScanned))
	for org := range orgScanned {
		out += fmt.Sprintf("	- %s\n", org)
	}

	out += fmt.Sprintf("Repositories Scanned: %d\n", len(repoScanned))
	for repo := range repoScanned {
		out += fmt.Sprintf("	- %s\n", repo)
	}

	out += fmt.Sprintf("Workflows Scanned: %d\n", len(workflowScanned))
	for workflow := range workflowScanned {
		out += fmt.Sprintf("	- %s\n", workflow)
	}

	out += fmt.Sprintf("Workflow Runs Scanned: %d\n", len(runsScanned))
	out += fmt.Sprintf("Secrets Found: %d\n", secretsFound)
	out += "===============================\n"

	fmt.Println(out)
}
