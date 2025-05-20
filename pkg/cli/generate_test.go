package cli

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/saasuke-labs/gengo/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := NewGenerateCommand()

	cmd.SetArgs([]string{"--manifest", "../../test-resources/simple-blog/input/gengo.yaml", "--output", "../../test-resources/simple-blog/output", "--plain"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

}

func prepareDirectories(testName string) (string, string, string) {

	base, err := filepath.Abs(path.Join("../../test-resources", testName))

	if err != nil {
		panic(err)
	}
	absInput := path.Join(base, "input")
	absOutput := path.Join(base, "output")
	absExpectedOutput := path.Join(base, "expected-output")

	return absInput, absOutput, absExpectedOutput
}
func TestGeneration(t *testing.T) {

	absInput, absOutput, absExpectedOutput := prepareDirectories("simple-blog")
	os.RemoveAll(absOutput)

	SilentGenerate([]string{path.Join(absInput, "gengo.yaml")}, absOutput)

	if _, err := os.Stat(absOutput); os.IsNotExist(err) {
		t.Fatalf("Output directory was not created: %v", err)
	}

	assert.FileExists(t, path.Join(absOutput, "blog/blog1.html"))

	assert.True(t, testutils.CompareHtmlFiles(
		path.Join(absOutput, "blog/blog1.html"),
		path.Join(absExpectedOutput, "blog/blog1.html"),
	))

}
