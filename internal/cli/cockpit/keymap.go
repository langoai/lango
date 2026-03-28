package cockpit

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ToggleSidebar key.Binding
	ToggleContext key.Binding
	FocusToggle   key.Binding
	CopyClipboard key.Binding
	Page1         key.Binding
	Page2         key.Binding
	Page3         key.Binding
	Page4         key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		ToggleSidebar: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "toggle sidebar"),
		),
		ToggleContext: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "toggle context panel"),
		),
		FocusToggle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		CopyClipboard: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy to clipboard"),
		),
		Page1: key.NewBinding(
			key.WithKeys("ctrl+1"),
			key.WithHelp("ctrl+1", "chat"),
		),
		Page2: key.NewBinding(
			key.WithKeys("ctrl+2"),
			key.WithHelp("ctrl+2", "settings"),
		),
		Page3: key.NewBinding(
			key.WithKeys("ctrl+3"),
			key.WithHelp("ctrl+3", "tools"),
		),
		Page4: key.NewBinding(
			key.WithKeys("ctrl+4"),
			key.WithHelp("ctrl+4", "status"),
		),
	}
}
