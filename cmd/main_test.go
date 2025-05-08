package main

import (
	"testing"
)

func TestIntegrationGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := NewGenerateCommand()

	cmd.SetArgs([]string{"--manifest", "../test-resources/simple-blog/input/gengo.yaml", "--output", "../test-resources/simple-blog/output", "--plain"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

}
