package cockpit

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ToggleSidebar key.Binding
	FocusToggle   key.Binding
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
		FocusToggle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
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
