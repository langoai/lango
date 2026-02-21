package types

// Enum defines a common interface for typed enum values.
// Each enum type should implement Valid() and Values() as value receiver methods.
type Enum[T any] interface {
	Valid() bool
	Values() []T
}
