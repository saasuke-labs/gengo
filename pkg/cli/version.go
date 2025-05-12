package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tonitienda/gengo/pkg/version"
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
