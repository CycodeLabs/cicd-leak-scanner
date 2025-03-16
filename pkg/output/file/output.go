package file

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileOutput struct {
	filename string
}

func New(filename string) (*FileOutput, error) {
	return &FileOutput{filename: filename}, nil
}

func (f *FileOutput) Write(org string, repo string, workflow string, secret string) error {
	var results map[string]map[string][]string

	if _, err := os.Stat(f.filename); err == nil {
		// File exists; open it.
		file, err := os.Open(f.filename)
		if err != nil {
			return fmt.Errorf("Failed to open file: %v", err)
		}
		defer file.Close()

		// Decode the existing JSON into the results map.
		if err := json.NewDecoder(file).Decode(&results); err != nil {
			return fmt.Errorf("Failed to decode JSON: %v", err)
		}
	} else {
		// File does not exist, initialize a new map.
		results = make(map[string]map[string][]string)
	}

	identifier := fmt.Sprintf("%s/%s", org, repo)

	if _, ok := results[identifier]; !ok {
		results[identifier] = make(map[string][]string)
	}

	// Append the secret entry to the workflowName list.
	// If the workflowName key doesn't exist, append will initialize the slice.
	results[identifier][workflow] = append(results[identifier][workflow], secret)

	// Open the file for writing (this creates or truncates the file).
	file, err := os.Create(f.filename)
	if err != nil {
		return fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Write the updated results to the file using a JSON encoder.
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("Failed to encode JSON: %v", err)
	}

	return nil
}
