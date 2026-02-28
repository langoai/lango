// Package tuicore provides shared TUI form components for CLI commands.
package tuicore

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

// InputType defines the type of input field.
type InputType int

const (
	InputText InputType = iota
	InputInt
	InputBool // Toggled via space
	InputSelect
	InputPassword
	InputSearchSelect // Searchable dropdown select
)

// Field represents a single configuration field in a form.
type Field struct {
	Key         string
	Label       string
	Description string // Help text shown below the focused field
	Type        InputType
	Value       string
	Placeholder string
	Options     []string // For InputSelect and InputSearchSelect
	Checked     bool     // For InputBool
	Width       int
	Validate    func(string) error

	// VisibleWhen controls conditional visibility. When non-nil, the field is
	// shown only when this function returns true. When nil the field is always visible.
	VisibleWhen func() bool

	// TextInput holds the bubbletea text input model (exported for cross-package use).
	TextInput textinput.Model
	Err       error

	// InputSearchSelect state
	FilteredOptions []string // Filtered subset of Options
	SelectCursor    int      // Cursor position in filtered list
	SelectOpen      bool     // Whether dropdown is open
}

// applySearchFilter filters Options by case-insensitive substring match.
func (f *Field) applySearchFilter(query string) {
	if query == "" {
		f.FilteredOptions = make([]string, len(f.Options))
		copy(f.FilteredOptions, f.Options)
	} else {
		q := strings.ToLower(query)
		f.FilteredOptions = f.FilteredOptions[:0]
		for _, opt := range f.Options {
			if strings.Contains(strings.ToLower(opt), q) {
				f.FilteredOptions = append(f.FilteredOptions, opt)
			}
		}
	}
	if f.SelectCursor >= len(f.FilteredOptions) {
		f.SelectCursor = max(0, len(f.FilteredOptions)-1)
	}
}

// IsVisible reports whether this field should be rendered and navigable.
func (f *Field) IsVisible() bool {
	if f.VisibleWhen == nil {
		return true
	}
	return f.VisibleWhen()
}
