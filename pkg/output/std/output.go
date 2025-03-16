package std

import "fmt"

type StdOutput struct{}

func New() (*StdOutput, error) {
	return &StdOutput{}, nil
}

func (s *StdOutput) Write(org string, repo string, workflow string, secret string) error {
	output := fmt.Sprintf("=========== %s/%s ===========\n", org, repo)
	output += fmt.Sprintf("Workflow: %s\n", workflow)
	output += fmt.Sprintf("Secret: %s\n", secret)
	output += "=============================\n"

	fmt.Println(output)
	return nil
}
