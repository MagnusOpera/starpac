package safety

import "testing"

func TestValidateDestructivePlan(t *testing.T) {
	if err := ValidateDestructivePlan(true, false, false, false); err == nil {
		t.Fatal("destructive plan was accepted without authorization")
	}
	for _, allowed := range [][3]bool{{true, false, false}, {false, true, false}, {false, false, true}} {
		if err := ValidateDestructivePlan(true, allowed[0], allowed[1], allowed[2]); err != nil {
			t.Fatalf("authorized destructive plan was rejected: %v", err)
		}
	}
}
