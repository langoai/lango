// Package tuicore provides shared TUI form components for CLI commands.
package tuicore

import "github.com/charmbracelet/bubbles/textinput"

// InputType defines the type of input field.
type InputType int

const (
	InputText InputType = iota
	InputInt
	InputBool // Toggled via space
	InputSelect
	InputPassword
)

// Field represents a single configuration field in a form.
type Field struct {
	Key         string
	Label       string
	Description string // Help text shown below the focused field
	Type        InputType
	Value       string
	Placeholder string
	Options     []string // For InputSelect
	Checked     bool     // For InputBool
	Width       int
	Validate    func(string) error

	// TextInput holds the bubbletea text input model (exported for cross-package use).
	TextInput textinput.Model
	Err       error
}
