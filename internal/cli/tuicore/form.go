package tuicore

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
)

// FormModel manages a list of fields.
type FormModel struct {
	Title    string
	Fields   []*Field
	Cursor   int
	Focus    bool
	OnSave   func(map[string]interface{})
	OnCancel func()
}

// NewFormModel creates a new form with the given title.
func NewFormModel(title string) FormModel {
	return FormModel{
		Title:  title,
		Fields: []*Field{},
		Cursor: 0,
	}
}

// AddField adds a field to the form, initializing its text input if applicable.
func (m *FormModel) AddField(f *Field) {
	if f.Type == InputText || f.Type == InputInt || f.Type == InputPassword {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.SetValue(f.Value)
		ti.CharLimit = 100
		ti.Width = 30
		if f.Width > 0 {
			ti.Width = f.Width
		}
		if f.Type == InputPassword {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '*'
		}
		f.TextInput = ti
	}
	m.Fields = append(m.Fields, f)
}

// Init implements tea.Model.
func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	if !m.Focus {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "tab":
			if m.Cursor < len(m.Fields)-1 {
				m.Cursor++
			}
		case " ":
			field := m.Fields[m.Cursor]
			if field.Type == InputBool {
				field.Checked = !field.Checked
			}
		case "esc":
			if m.OnCancel != nil {
				m.OnCancel()
			}
		}
	}

	// Update specific field logic
	field := m.Fields[m.Cursor]
	if field.Type == InputText || field.Type == InputInt || field.Type == InputPassword {
		var inputCmd tea.Cmd
		field.TextInput, inputCmd = field.TextInput.Update(msg)
		field.Value = field.TextInput.Value()
		cmd = inputCmd
	}

	// Handle Select Logic (Left/Right to cycle options)
	if field.Type == InputSelect {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "right", "l":
				idx := -1
				for i, opt := range field.Options {
					if opt == field.Value {
						idx = i
						break
					}
				}
				if idx < len(field.Options)-1 {
					field.Value = field.Options[idx+1]
				} else if len(field.Options) > 0 {
					field.Value = field.Options[0]
				}
			case "left", "h":
				idx := -1
				for i, opt := range field.Options {
					if opt == field.Value {
						idx = i
						break
					}
				}
				if idx > 0 {
					field.Value = field.Options[idx-1]
				} else if len(field.Options) > 0 {
					field.Value = field.Options[len(field.Options)-1]
				}
			}
		}
	}

	return m, cmd
}

// View renders the form.
func (m FormModel) View() string {
	var b strings.Builder

	b.WriteString(tui.FormTitleBarStyle.Render(m.Title))
	b.WriteString("\n")

	for i, f := range m.Fields {
		labelStyle := lipgloss.NewStyle().Width(20)
		if i == m.Cursor {
			labelStyle = labelStyle.Foreground(tui.Accent).Bold(true)
		}

		b.WriteString(labelStyle.Render(f.Label))

		switch f.Type {
		case InputText, InputInt, InputPassword:
			if i == m.Cursor {
				f.TextInput.Focus()
				f.TextInput.TextStyle = lipgloss.NewStyle().Foreground(tui.Accent)
			} else {
				f.TextInput.Blur()
				f.TextInput.TextStyle = lipgloss.NewStyle()
			}
			b.WriteString(f.TextInput.View())

		case InputBool:
			check := "[ ]"
			if f.Checked {
				check = "[x]"
			}
			if i == m.Cursor {
				check = lipgloss.NewStyle().Foreground(tui.Accent).Render(check)
			}
			b.WriteString(check)

		case InputSelect:
			val := f.Value
			if val == "" && len(f.Options) > 0 {
				val = f.Options[0]
			}
			if i == m.Cursor {
				val = fmt.Sprintf("< %s >", val)
				val = lipgloss.NewStyle().Foreground(tui.Accent).Render(val)
			}
			b.WriteString(val)
		}
		b.WriteString("\n")
	}

	// Help Footer
	b.WriteString("\n")
	b.WriteString(tui.HelpBar(
		tui.HelpEntry("Tab", "Next"),
		tui.HelpEntry("Shift+Tab", "Prev"),
		tui.HelpEntry("Space", "Toggle"),
		tui.HelpEntry("←→", "Options"),
		tui.HelpEntry("Esc", "Back"),
	))

	return b.String()
}
