package artisan

import (
	"errors"
	"fmt"
	"strings"
)

// Run executes the artisan command pipeline.
func Run(args []string) error {
	expanded := expandColonArgs(args)
	if len(expanded) == 0 {
		return errors.New("no artisan command provided")
	}

	cmd := expanded[0]
	rest := expanded[1:]

	switch cmd {
	case "migrate":
		return runMigrate(rest)
	case "make":
		if len(rest) == 0 {
			return errors.New("please specify what to make (e.g. make model Foo)")
		}
		return runMake(rest)
	default:
		return fmt.Errorf("unknown artisan command: %s", strings.Join(expanded, " "))
	}
}

func runMake(args []string) error {
	target := args[0]
	rest := args[1:]

	switch target {
	case "model":
		return runMakeModel(rest)
	case "controller":
		return runMakeController(rest)
	case "repository":
		return runMakeRepository(rest)
	default:
		return fmt.Errorf("unknown make target: %s", target)
	}
}

func expandColonArgs(args []string) []string {
	var expanded []string
	for _, arg := range args {
		parts := strings.Split(arg, ":")
		for _, part := range parts {
			if part != "" {
				expanded = append(expanded, part)
			}
		}
		if len(parts) == 0 {
			expanded = append(expanded, arg)
		}
	}
	return expanded
}
