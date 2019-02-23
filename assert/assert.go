package assert

import "fmt"

func EqualString(expected string, got string) error {
	if expected != got {
		return fmt.Errorf("expected:\n%q\ngot\n%q", expected, got)
	}
	return nil
}
