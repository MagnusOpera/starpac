package safety

import "fmt"

func ValidateDestructivePlan(destructive, configured, allowDrop, force bool) error {
	if destructive && !(configured || allowDrop || force) {
		return fmt.Errorf("plan contains destructive operations; re-run with --allow-drop or --force")
	}
	return nil
}
