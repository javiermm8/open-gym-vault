package api

import "fmt"

func errRequired(field string) error {
	return fmt.Errorf("%s is required", field)
}

func errInvalid(message string) error {
	return fmt.Errorf("%s", message)
}
