package cli

import (
	"fmt"

	"github.com/saasuke-labs/gengo/pkg/version"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", version.Version, version.Commit, version.Date)
		},
	}

	return versionCmd
}
