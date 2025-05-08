package cli

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

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

	absInput, absOutput, _ := prepareDirectories("simple-blog")
	os.RemoveAll(absOutput)

	SilentGenerate(path.Join(absInput, "gengo.yaml"), absOutput)

	if _, err := os.Stat(absOutput); os.IsNotExist(err) {
		t.Fatalf("Output directory was not created: %v", err)
	}

}
