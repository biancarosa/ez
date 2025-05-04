package ez

import "fmt"

func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkError(err error) error {
	if err != nil {
		return fmt.Errorf("ez operation failed: %w", err)
	}
	return nil
}
