// Package configcmd provides lango config get/set/keys subcommands.
package configcmd

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/cli/clihttp"
	"github.com/langoai/lango/internal/config"
)

// NewGetCmd creates the "config get <dot.path>" command.
func NewGetCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "get <dot.path>",
		Short:         "Read a configuration value by dot-notation path",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Read a configuration value using dot-notation (e.g. agent.provider, p2p.enabled).

This is a read-only operation. Use "lango config set" to modify values.

Examples:
  lango config get agent.provider
  lango config get p2p.enabled
  lango config get economy.budget.defaultMax
  lango config get agent --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFmt, err := resolvePlainOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			val, err := resolveConfigPath(cfg, args[0])
			if err != nil {
				return err
			}

			return printValue(cmd.OutOrStdout(), val, outputFmt)
		},
	}

	cmd.Flags().StringP("output", "o", "plain", "Output format (plain, json)")
	return cmd
}

func resolvePlainOrJSONOutput(cmd *cobra.Command) (string, error) {
	flag, _ := cmd.Flags().GetString("output")
	switch normalized := strings.ToLower(strings.TrimSpace(flag)); normalized {
	case "", "plain":
		return "plain", nil
	case "json":
		return "json", nil
	default:
		return "", fmt.Errorf("unknown output format %q (expected: plain or json)", strings.TrimSpace(flag))
	}
}

// NewSetCmd creates the "config set <dot.path> <value>" command.
// The passphrase is implicitly verified via bootstrap (caller must bootstrap first).
// cfgLoader returns (config, explicitKeys, cleanup, error). cleanup closes
// bootstrap resources and is called via defer in RunE so resources are released
// on all code paths.
func NewSetCmd(
	cfgLoader func() (*config.Config, map[string]bool, func(), error),
	cfgSaver func(*config.Config, map[string]bool) error,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <dot.path> <value>",
		Short: "Set a configuration value (requires passphrase verification)",
		Long: `Set a configuration value using dot-notation.

This command requires passphrase verification because it modifies the encrypted
configuration profile. AI agents calling this command will be prompted for the
passphrase interactively, preventing unauthorized config changes.

Examples:
  lango config set agent.provider openai
  lango config set p2p.enabled true
  lango config set economy.budget.defaultMax 20.00`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, explicitKeys, cleanup, err := cfgLoader()
			if cleanup != nil {
				defer cleanup()
			}
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if err := setConfigPath(cfg, args[0], args[1]); err != nil {
				return err
			}

			explicitKeys = explicitKeysForSetPath(explicitKeys, args[0])
			if err := cfgSaver(cfg, explicitKeys); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", args[0], args[1])
			return err
		},
	}

	return cmd
}

func explicitKeysForSetPath(existing map[string]bool, path string) map[string]bool {
	isContextRelated := false
	for _, key := range config.ContextRelatedKeys() {
		if key == path {
			isContextRelated = true
			break
		}
	}
	if existing == nil && !isContextRelated {
		return nil
	}

	out := make(map[string]bool, len(existing)+1)
	for key, value := range existing {
		out[key] = value
	}
	if isContextRelated {
		out[path] = true
	}
	return out
}

// NewKeysCmd creates the "config keys [prefix]" command.
func NewKeysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keys [prefix]",
		Short: "List available configuration keys",
		Long: `List available configuration keys using mapstructure tags.

Optionally filter by a dot-path prefix.

Examples:
  lango config keys
  lango config keys agent
  lango config keys p2p.zkp`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}

			keys := collectKeys(reflect.TypeOf(config.Config{}), "")
			sort.Strings(keys)

			for _, k := range keys {
				if prefix == "" || strings.HasPrefix(k, prefix) {
					fmt.Fprintln(cmd.OutOrStdout(), k)
				}
			}

			return nil
		},
	}
}

// resolveConfigPath traverses the config struct using dot-notation and mapstructure tags.
func resolveConfigPath(cfg *config.Config, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(cfg).Elem()
	prefixParts := make([]string, 0, len(parts))

	for _, part := range parts {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil, fmt.Errorf("config path %q: nil pointer at %q", path, part)
			}
			v = v.Elem()
		}

		if v.Kind() == reflect.Map {
			mapKey := reflect.ValueOf(part)
			mv := v.MapIndex(mapKey)
			if !mv.IsValid() {
				return nil, fmt.Errorf("config path %q: key %q not found in map", path, part)
			}
			v = mv
			continue
		}

		if v.Kind() != reflect.Struct {
			return nil, nonStructConfigPathError(
				path,
				part,
				v.Kind().String(),
				strings.Join(prefixParts, "."),
			)
		}

		idx := findFieldByTag(v.Type(), part)
		if idx < 0 {
			return nil, unknownConfigFieldError(path, part, strings.Join(prefixParts, "."))
		}
		v = v.Field(idx)
		prefixParts = append(prefixParts, part)
	}

	return v.Interface(), nil
}

// setConfigPath traverses the config struct and sets a value at the given path.
func setConfigPath(cfg *config.Config, path, rawVal string) error {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(cfg).Elem()
	prefixParts := make([]string, 0, len(parts))

	for i, part := range parts {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return nonStructConfigPathError(
				path,
				part,
				"",
				strings.Join(prefixParts, "."),
			)
		}

		idx := findFieldByTag(v.Type(), part)
		if idx < 0 {
			return unknownConfigFieldError(path, part, strings.Join(prefixParts, "."))
		}

		if i < len(parts)-1 {
			v = v.Field(idx)
			prefixParts = append(prefixParts, part)
			continue
		}

		// Last segment — set the value.
		field := v.Field(idx)
		return setField(field, rawVal, path)
	}

	return fmt.Errorf("config path %q: empty path", path)
}

func unknownConfigFieldError(path, field, validPrefix string) error {
	return configPathDiscoveryError(
		fmt.Sprintf("config path %q: field %q not found", path, field),
		validPrefix,
		suggestConfigKeys(path, validPrefix),
	)
}

func nonStructConfigPathError(path, field, kind, leafPath string) error {
	message := fmt.Sprintf("config path %q: %q is not a struct", path, field)
	if kind != "" {
		message += fmt.Sprintf(" (kind: %s)", kind)
	}
	return configPathDiscoveryError(
		message,
		parentConfigPrefix(leafPath),
		[]string{leafPath},
	)
}

func configPathDiscoveryError(message, validPrefix string, suggestions []string) error {
	parts := []string{message}
	if suggestions = uniqueNonEmptyStrings(suggestions); len(suggestions) > 0 {
		parts = append(parts, "did you mean: "+strings.Join(suggestions, ", "))
	}

	hint := "lango config keys"
	if validPrefix != "" {
		hint += " " + validPrefix
	}
	parts = append(parts, "list keys: "+hint)

	return fmt.Errorf("%s", strings.Join(parts, "; "))
}

func parentConfigPrefix(path string) string {
	if idx := strings.LastIndex(path, "."); idx > 0 {
		return path[:idx]
	}
	return ""
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func suggestConfigKeys(path, validPrefix string) []string {
	keys := collectKeys(reflect.TypeOf(config.Config{}), "")
	samePrefix := make([]string, 0, len(keys))
	if validPrefix != "" {
		prefix := validPrefix + "."
		for _, key := range keys {
			if strings.HasPrefix(key, prefix) {
				samePrefix = append(samePrefix, key)
			}
		}
	}
	if len(samePrefix) > 0 {
		return nearestConfigKeys(path, samePrefix)
	}
	return nearestConfigKeys(path, keys)
}

func nearestConfigKeys(path string, keys []string) []string {
	type candidate struct {
		key      string
		distance int
	}

	candidates := make([]candidate, 0, len(keys))
	for _, key := range keys {
		distance := editDistance(path, key)
		if distance <= 3 {
			candidates = append(candidates, candidate{key: key, distance: distance})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].key < candidates[j].key
	})

	limit := len(candidates)
	if limit > 3 {
		limit = 3
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, candidates[i].key)
	}
	return out
}

func editDistance(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return len(b)
	}
	if b == "" {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = minInt(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func minInt(first int, rest ...int) int {
	minimum := first
	for _, value := range rest {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}

// setField sets a reflect.Value from a raw string based on its type.
func setField(field reflect.Value, rawVal, path string) error {
	if field.Kind() == reflect.Ptr {
		elem := reflect.New(field.Type().Elem())
		if err := setField(elem.Elem(), rawVal, path); err != nil {
			return err
		}
		field.Set(elem)
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(rawVal)
	case reflect.Bool:
		b, err := strconv.ParseBool(rawVal)
		if err != nil {
			return fmt.Errorf("config path %q: invalid bool %q", path, rawVal)
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int64:
		// Handle time.Duration (int64 nanoseconds)
		if field.Type().String() == "time.Duration" {
			return fmt.Errorf("config path %q: duration fields should use 'lango settings' TUI", path)
		}
		i, err := strconv.ParseInt(rawVal, 10, 64)
		if err != nil {
			return fmt.Errorf("config path %q: invalid integer %q", path, rawVal)
		}
		field.SetInt(i)
	case reflect.Uint64:
		u, err := strconv.ParseUint(rawVal, 10, 64)
		if err != nil {
			return fmt.Errorf("config path %q: invalid unsigned integer %q", path, rawVal)
		}
		field.SetUint(u)
	case reflect.Float64:
		f, err := strconv.ParseFloat(rawVal, 64)
		if err != nil {
			return fmt.Errorf("config path %q: invalid float %q", path, rawVal)
		}
		field.SetFloat(f)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(rawVal, ",")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				if t := strings.TrimSpace(p); t != "" {
					out = append(out, t)
				}
			}
			field.Set(reflect.ValueOf(out))
		} else {
			return fmt.Errorf("config path %q: unsupported slice type", path)
		}
	default:
		return fmt.Errorf("config path %q: unsupported type %s (use 'lango settings' for complex fields)", path, field.Kind())
	}
	return nil
}

// findFieldByTag finds a struct field index by its mapstructure tag value.
func findFieldByTag(t reflect.Type, tag string) int {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ms := f.Tag.Get("mapstructure")
		if ms == tag {
			return i
		}
	}
	return -1
}

// collectKeys recursively collects all leaf config keys using mapstructure tags.
func collectKeys(t reflect.Type, prefix string) []string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var keys []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		fullKey := tag
		if prefix != "" {
			fullKey = prefix + "." + tag
		}

		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct && ft.String() != "time.Duration" {
			// Skip map types (providers, servers, etc.)
			if f.Type.Kind() == reflect.Map {
				keys = append(keys, fullKey+".<name>.*")
				continue
			}
			keys = append(keys, collectKeys(ft, fullKey)...)
		} else {
			keys = append(keys, fullKey)
		}
	}
	return keys
}

// printValue formats and prints a value.
func printValue(w io.Writer, val interface{}, format string) error {
	if format == "json" {
		if err := clihttp.PrintJSON(w, val); err != nil {
			return fmt.Errorf("marshal value: %w", err)
		}
		return nil
	}

	// plain format
	_, err := fmt.Fprintln(w, formatPlain(val))
	return err
}

// formatPlain converts a value to a human-readable string.
func formatPlain(val interface{}) string {
	if val == nil {
		return "<nil>"
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return "<nil>"
		}
		return formatPlain(rv.Elem().Interface())
	case reflect.Slice:
		parts := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts[i] = fmt.Sprintf("%v", rv.Index(i).Interface())
		}
		return strings.Join(parts, ",")
	case reflect.Map:
		parts := make([]string, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			parts = append(parts, fmt.Sprintf("%v=%v", k.Interface(), rv.MapIndex(k).Interface()))
		}
		sort.Strings(parts)
		return strings.Join(parts, ",")
	default:
		return fmt.Sprintf("%v", val)
	}
}
