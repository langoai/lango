package extension

import (
	"io"

	"github.com/langoai/lango/internal/cli/clihttp"
)

func printJSON(w io.Writer, v interface{}) error {
	return clihttp.PrintJSON(w, v)
}
