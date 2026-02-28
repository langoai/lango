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
	Cursor   int // index into VisibleFields()
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
	if f.Type == InputSearchSelect {
		ti := textinput.New()
		ti.Placeholder = "Type to search..."
		ti.SetValue(f.Value)
		ti.CharLimit = 200
		ti.Width = 40
		if f.Width > 0 {
			ti.Width = f.Width
		}
		f.TextInput = ti
		f.FilteredOptions = make([]string, len(f.Options))
		copy(f.FilteredOptions, f.Options)
	}
	m.Fields = append(m.Fields, f)
}

// HasOpenDropdown reports whether any field has an open search-select dropdown.
func (m FormModel) HasOpenDropdown() bool {
	for _, f := range m.Fields {
		if f.Type == InputSearchSelect && f.SelectOpen {
			return true
		}
	}
	return false
}

// VisibleFields returns only the fields that pass their visibility check.
func (m FormModel) VisibleFields() []*Field {
	var out []*Field
	for _, f := range m.Fields {
		if f.IsVisible() {
			out = append(out, f)
		}
	}
	return out
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

	visible := m.VisibleFields()
	if len(visible) == 0 {
		return m, nil
	}

	// Clamp cursor in case visibility changed.
	if m.Cursor >= len(visible) {
		m.Cursor = len(visible) - 1
	}

	var cmd tea.Cmd

	field := visible[m.Cursor]

	// InputSearchSelect with open dropdown: intercept keys before form navigation.
	if field.Type == InputSearchSelect && field.SelectOpen {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "up":
				if field.SelectCursor > 0 {
					field.SelectCursor--
				}
				return m, nil
			case "down":
				if field.SelectCursor < len(field.FilteredOptions)-1 {
					field.SelectCursor++
				}
				return m, nil
			case "enter":
				if len(field.FilteredOptions) > 0 && field.SelectCursor < len(field.FilteredOptions) {
					field.Value = field.FilteredOptions[field.SelectCursor]
					field.TextInput.SetValue(field.Value)
				}
				field.SelectOpen = false
				return m, nil
			case "esc":
				field.SelectOpen = false
				field.TextInput.SetValue(field.Value)
				field.applySearchFilter("")
				return m, nil
			case "tab":
				field.SelectOpen = false
				field.TextInput.SetValue(field.Value)
				field.applySearchFilter("")
				if m.Cursor < len(visible)-1 {
					m.Cursor++
				}
				return m, nil
			case "shift+tab":
				field.SelectOpen = false
				field.TextInput.SetValue(field.Value)
				field.applySearchFilter("")
				if m.Cursor > 0 {
					m.Cursor--
				}
				return m, nil
			default:
				// Pass character input to text field for filtering
				var inputCmd tea.Cmd
				field.TextInput, inputCmd = field.TextInput.Update(msg)
				field.applySearchFilter(field.TextInput.Value())
				cmd = inputCmd
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "tab":
			if m.Cursor < len(visible)-1 {
				m.Cursor++
			}
		case " ":
			if field.Type == InputBool {
				field.Checked = !field.Checked
			}
		case "enter":
			if field.Type == InputSearchSelect {
				field.SelectOpen = true
				field.SelectCursor = 0
				field.TextInput.SetValue("")
				field.applySearchFilter("")
				field.TextInput.Focus()
				// Pre-select current value in list
				for i, opt := range field.FilteredOptions {
					if opt == field.Value {
						field.SelectCursor = i
						break
					}
				}
				return m, nil
			}
		case "esc":
			if m.OnCancel != nil {
				m.OnCancel()
			}
		}
	}

	// Re-evaluate visible after potential toggle change.
	visible = m.VisibleFields()
	if len(visible) == 0 {
		return m, nil
	}
	if m.Cursor >= len(visible) {
		m.Cursor = len(visible) - 1
	}

	// Update specific field logic.
	field = visible[m.Cursor]
	if field.Type == InputText || field.Type == InputInt || field.Type == InputPassword {
		var inputCmd tea.Cmd
		field.TextInput, inputCmd = field.TextInput.Update(msg)
		field.Value = field.TextInput.Value()
		cmd = inputCmd
	}

	// Handle Select Logic (Left/Right to cycle options).
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

	visible := m.VisibleFields()
	for vi, f := range visible {
		isFocused := vi == m.Cursor

		labelStyle := lipgloss.NewStyle().Width(20)
		if isFocused {
			labelStyle = labelStyle.Foreground(tui.Accent).Bold(true)
		}

		b.WriteString(labelStyle.Render(f.Label))

		switch f.Type {
		case InputText, InputInt, InputPassword:
			if isFocused {
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
			if isFocused {
				check = lipgloss.NewStyle().Foreground(tui.Accent).Render(check)
			}
			b.WriteString(check)

		case InputSelect:
			val := f.Value
			if val == "" && len(f.Options) > 0 {
				val = f.Options[0]
			}
			if isFocused {
				val = fmt.Sprintf("< %s >", val)
				val = lipgloss.NewStyle().Foreground(tui.Accent).Render(val)
			}
			b.WriteString(val)

		case InputSearchSelect:
			if isFocused && f.SelectOpen {
				// Show search input
				f.TextInput.Focus()
				f.TextInput.TextStyle = lipgloss.NewStyle().Foreground(tui.Accent)
				b.WriteString(f.TextInput.View())
				b.WriteString("\n")

				// Show match count
				matchInfo := fmt.Sprintf("  %d/%d matches", len(f.FilteredOptions), len(f.Options))
				b.WriteString(lipgloss.NewStyle().Foreground(tui.Dim).Render(matchInfo))
				b.WriteString("\n")

				// Render dropdown (max 8 visible)
				maxVisible := 8
				start := 0
				if f.SelectCursor >= maxVisible {
					start = f.SelectCursor - maxVisible + 1
				}
				end := start + maxVisible
				if end > len(f.FilteredOptions) {
					end = len(f.FilteredOptions)
				}

				for i := start; i < end; i++ {
					opt := f.FilteredOptions[i]
					if i == f.SelectCursor {
						b.WriteString(lipgloss.NewStyle().Foreground(tui.Accent).Bold(true).Render("  > " + opt))
					} else {
						b.WriteString(lipgloss.NewStyle().Foreground(tui.Muted).Render("    " + opt))
					}
					b.WriteString("\n")
				}

				if end < len(f.FilteredOptions) {
					more := fmt.Sprintf("    ... %d more", len(f.FilteredOptions)-end)
					b.WriteString(lipgloss.NewStyle().Foreground(tui.Dim).Render(more))
					b.WriteString("\n")
				}
			} else {
				// Closed state: show current value
				val := f.Value
				if val == "" {
					val = "(none)"
				}
				if isFocused {
					val = lipgloss.NewStyle().Foreground(tui.Accent).Render(val + "  [Enter: search]")
				}
				b.WriteString(val)
			}
		}
		b.WriteString("\n")

		// Show description for the focused field.
		if isFocused && f.Description != "" {
			b.WriteString(tui.FieldDescStyle.Render("ℹ " + f.Description))
			b.WriteString("\n")
		}
	}

	// Help Footer - context-dependent
	b.WriteString("\n")
	hasOpenDropdown := false
	for _, f := range visible {
		if f.Type == InputSearchSelect && f.SelectOpen {
			hasOpenDropdown = true
			break
		}
	}
	if hasOpenDropdown {
		b.WriteString(tui.HelpBar(
			tui.HelpEntry("↑↓", "Navigate"),
			tui.HelpEntry("Enter", "Select"),
			tui.HelpEntry("Esc", "Close"),
			tui.HelpEntry("Type", "Filter"),
		))
	} else {
		b.WriteString(tui.HelpBar(
			tui.HelpEntry("Tab", "Next"),
			tui.HelpEntry("Shift+Tab", "Prev"),
			tui.HelpEntry("Space", "Toggle"),
			tui.HelpEntry("←→", "Options"),
			tui.HelpEntry("Enter", "Search"),
			tui.HelpEntry("Esc", "Back"),
		))
	}

	return b.String()
}
