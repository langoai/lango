package configstore

import "time"

// ProfileInfo holds metadata about a configuration profile.
type ProfileInfo struct {
	Name      string
	Active    bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
