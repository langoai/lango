package tuicore

import (
	"fmt"
	"strconv"
)

// TextInput creates a text input field with the given key, label, value, and description.
func TextInput(key, label, value, desc string) *Field {
	return &Field{Key: key, Label: label, Type: InputText, Value: value, Description: desc}
}

// TextInputWithPlaceholder creates a text input field with a placeholder hint.
func TextInputWithPlaceholder(key, label, value, placeholder, desc string) *Field {
	return &Field{Key: key, Label: label, Type: InputText, Value: value, Placeholder: placeholder, Description: desc}
}

// PasswordInput creates a password input field.
func PasswordInput(key, label, value, desc string) *Field {
	return &Field{Key: key, Label: label, Type: InputPassword, Value: value, Description: desc}
}

// IntInput creates an integer input field with positive-integer validation.
func IntInput(key, label string, value int, desc string) *Field {
	return &Field{
		Key: key, Label: label, Type: InputInt,
		Value:       strconv.Itoa(value),
		Description: desc,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	}
}

// BoolInput creates a boolean toggle field.
func BoolInput(key, label string, checked bool, desc string) *Field {
	return &Field{Key: key, Label: label, Type: InputBool, Checked: checked, Description: desc}
}

// SelectInput creates a select field with a list of options.
func SelectInput(key, label, value string, options []string, desc string) *Field {
	return &Field{Key: key, Label: label, Type: InputSelect, Value: value, Options: options, Description: desc}
}

// SearchSelectInput creates a searchable select field.
func SearchSelectInput(key, label, value string, options []string, desc string) *Field {
	return &Field{Key: key, Label: label, Type: InputSearchSelect, Value: value, Options: options, Description: desc}
}
