package cmd

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/spf13/cobra"
)

// NewCmdExtract creates a new cobra.Command for the extract subcommand.
func NewCmdExtract(options *[]crane.Option) *cobra.Command {
	var pattern string

	exportCmd := &cobra.Command{
		Use:   "extract IMAGE DIRECTORY",
		Short: "Extract contents of a remote image to a directory",
		Example: `  # Write content to a directory
  crane extract ubuntu ./

  # Write content to a directory that matches a pattern
  crane extract ubuntu ./ --pattern /usr/bin/*`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			src, dst := args[0], args[1]

			img, err := crane.Pull(src, *options...)
			if err != nil {
				return fmt.Errorf("pulling %s: %w", src, err)
			}
			return crane.Extract(img, crane.ExtractArgs{dst: pattern})
		},
	}
	exportCmd.Flags().StringVarP(&pattern, "pattern", "p", "*", "The shell file name pattern to export only a subset of files from an image. If nothing is based default behavior occurs.")
	return exportCmd
}
