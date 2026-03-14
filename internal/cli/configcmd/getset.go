// Package configcmd provides lango config get/set/keys subcommands.
package configcmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

// NewGetCmd creates the "config get <dot.path>" command.
func NewGetCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var outputFmt string

	cmd := &cobra.Command{
		Use:   "get <dot.path>",
		Short: "Read a configuration value by dot-notation path",
		Long: `Read a configuration value using dot-notation (e.g. agent.provider, p2p.enabled).

This is a read-only operation. Use "lango config set" to modify values.

Examples:
  lango config get agent.provider
  lango config get p2p.enabled
  lango config get economy.budget.defaultMax
  lango config get agent --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			val, err := resolveConfigPath(cfg, args[0])
			if err != nil {
				return err
			}

			return printValue(val, outputFmt)
		},
	}

	cmd.Flags().StringVarP(&outputFmt, "output", "o", "plain", "Output format (plain, json)")
	return cmd
}

// NewSetCmd creates the "config set <dot.path> <value>" command.
// The passphrase is implicitly verified via bootstrap (caller must bootstrap first).
func NewSetCmd(
	cfgLoader func() (*config.Config, error),
	cfgSaver func(*config.Config) error,
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
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if err := setConfigPath(cfg, args[0], args[1]); err != nil {
				return err
			}

			if err := cfgSaver(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("Set %s = %s\n", args[0], args[1])
			return nil
		},
	}

	return cmd
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
					fmt.Println(k)
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
			return nil, fmt.Errorf("config path %q: %q is not a struct (kind: %s)", path, part, v.Kind())
		}

		idx := findFieldByTag(v.Type(), part)
		if idx < 0 {
			return nil, fmt.Errorf("config path %q: field %q not found", path, part)
		}
		v = v.Field(idx)
	}

	return v.Interface(), nil
}

// setConfigPath traverses the config struct and sets a value at the given path.
func setConfigPath(cfg *config.Config, path, rawVal string) error {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(cfg).Elem()

	for i, part := range parts {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return fmt.Errorf("config path %q: %q is not a struct", path, part)
		}

		idx := findFieldByTag(v.Type(), part)
		if idx < 0 {
			return fmt.Errorf("config path %q: field %q not found", path, part)
		}

		if i < len(parts)-1 {
			v = v.Field(idx)
			continue
		}

		// Last segment — set the value.
		field := v.Field(idx)
		return setField(field, rawVal, path)
	}

	return fmt.Errorf("config path %q: empty path", path)
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
func printValue(val interface{}, format string) error {
	if format == "json" {
		data, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal value: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// plain format
	fmt.Println(formatPlain(val))
	return nil
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
