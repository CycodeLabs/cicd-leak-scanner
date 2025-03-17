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
		file, err := os.Open(f.filename)
		if err != nil {
			return fmt.Errorf("Failed to open file: %v", err)
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&results); err != nil {
			return fmt.Errorf("Failed to decode JSON: %v", err)
		}
	} else {
		results = make(map[string]map[string][]string)
	}

	identifier := fmt.Sprintf("%s/%s", org, repo)

	if _, ok := results[identifier]; !ok {
		results[identifier] = make(map[string][]string)
	}

	results[identifier][workflow] = append(results[identifier][workflow], secret)

	file, err := os.Create(f.filename)
	if err != nil {
		return fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("Failed to encode JSON: %v", err)
	}

	return nil
}
