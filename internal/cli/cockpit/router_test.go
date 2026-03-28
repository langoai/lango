package cockpit

import "testing"

func TestPageIDString(t *testing.T) {
	tests := []struct {
		give PageID
		want string
	}{
		{give: PageChat, want: "chat"},
		{give: PageSettings, want: "settings"},
		{give: PageTools, want: "tools"},
		{give: PageStatus, want: "status"},
		{give: PageID(99), want: "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.give.String(); got != tt.want {
				t.Errorf("PageID(%d).String() = %q, want %q",
					tt.give, got, tt.want)
			}
		})
	}
}

func TestPageIDFromString(t *testing.T) {
	tests := []struct {
		give string
		want PageID
	}{
		{give: "chat", want: PageChat},
		{give: "settings", want: PageSettings},
		{give: "tools", want: PageTools},
		{give: "status", want: PageStatus},
		{give: "unknown-value", want: PageChat},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			if got := PageIDFromString(tt.give); got != tt.want {
				t.Errorf("PageIDFromString(%q) = %d, want %d",
					tt.give, got, tt.want)
			}
		})
	}
}
