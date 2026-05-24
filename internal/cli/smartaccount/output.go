package smartaccount

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
)

func resolveTableOrJSONOutput(cmd *cobra.Command) (string, error) {
	return clihttp.ResolveTableOrJSONOutput(cmd)
}

func printJSON(w io.Writer, v interface{}) error {
	return clihttp.PrintJSON(w, v)
}
