package audit

import "testing"

func TestRunAudit(t *testing.T) {
	r := Run()
	if len(r.Checks) < 5 {
		t.Fatalf("expected several checks, got %d", len(r.Checks))
	}
	if r.Summary == "" {
		t.Fatal("empty summary")
	}
	t.Logf("canProceed=%v windows=%v node=%v", r.CanProceed, r.WindowsOK, r.NodeOK)
}
