package cli

import (
	"fmt"
	"os"

	"prune/internal/rules"
)

func runRules(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %v", args)
	}

	list := rules.All()
	for _, rule := range list {
		if _, err := fmt.Fprintf(os.Stdout, "%s\t%s\n", rule.ID, rule.Description); err != nil {
			return err
		}
	}
	return nil
}
