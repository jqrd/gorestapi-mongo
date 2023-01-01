package cmd

import (
	"fmt"

	cli "github.com/spf13/cobra"

	"github.com/jqrd/gorestapi-mongo/pkg/version"
)

// Version command
func init() {
	rootCmd.AddCommand(&cli.Command{
		Use:   "version",
		Short: "Show version",
		Long:  `Show version`,
		Run: func(cmd *cli.Command, args []string) {
			fmt.Println(version.Executable + " - " + version.GitVersion)
		},
	})
}
