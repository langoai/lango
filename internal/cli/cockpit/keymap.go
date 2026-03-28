package cockpit

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	ToggleSidebar key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		ToggleSidebar: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "toggle sidebar"),
		),
	}
}
