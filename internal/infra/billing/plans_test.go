package billing

import "testing"

func TestAllPlansIsStableAndIncludesFreePlan(t *testing.T) {
	all := AllPlans()

	if len(all) != 4 {
		t.Fatalf("expected 4 plans, got %d", len(all))
	}

	if all[0].Key != "edge" || all[1].Key != "free" {
		t.Fatalf("plans are not sorted: %+v", all)
	}

	if _, ok := PlanByKey("missing"); ok {
		t.Fatal("expected missing plan lookup to fail")
	}
}
