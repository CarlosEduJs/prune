package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"prune/internal/rules"
)

func runRules(_ context.Context, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %v", args)
	}

	list := rules.All()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "RULE ID\tDESCRIPTION")

	for _, rule := range list {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", rule.ID, rule.Description); err != nil {
			return err
		}
	}
	return w.Flush()
}
