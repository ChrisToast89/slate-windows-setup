package installstatus

import (
	"path/filepath"
	"testing"
)

func TestProbeUserNPMTree(t *testing.T) {
	// The working npm install the user runs from.
	root := `M:\Users\Chris\Documents\_code-projects\slate\slate`
	inst, ok := probeNPMTree(root)
	if !ok {
		// Tree may not exist on other machines
		if !fileExists(filepath.Join(root, "package.json")) {
			t.Skip("user slate npm tree not present")
		}
		t.Fatalf("expected to detect npm slate at %s", root)
	}
	if inst.Flavor != FlavorNPM {
		t.Fatalf("flavor %s", inst.Flavor)
	}
	if !inst.Healthy {
		t.Fatalf("expected healthy npm install, got %#v", inst)
	}
	t.Logf("detected: %s (%s)", inst.Root, inst.Description)
}

func TestIgnoreWailsTree(t *testing.T) {
	// Archived early Wails port (was Slate-win/slate-windows).
	root := `M:\Users\Chris\Documents\_code-projects\slate\_archive\early-wails-port-NOT-ACTIVE`
	if !fileExists(filepath.Join(root, "wails.json")) {
		t.Skip("archived early wails port not present")
	}
	if _, ok := probeNPMTree(root); ok {
		t.Fatal("must not treat archived wails port as npm slate")
	}
	if !isIgnoredTree(root) {
		t.Fatal("archived early wails port should be ignored")
	}
}

func TestDetectFindsNPM(t *testing.T) {
	root := `M:\Users\Chris\Documents\_code-projects\slate\slate`
	if !fileExists(filepath.Join(root, "package.json")) {
		t.Skip("no local npm slate")
	}
	st := Detect()
	found := false
	for _, inst := range st.Instances {
		if inst.Flavor == FlavorNPM && filepath.Clean(inst.Root) == filepath.Clean(root) {
			found = true
		}
	}
	if !found && !st.Installed {
		// Detect should at least mark installed if that tree is healthy
		t.Logf("status: %+v", st.Summary)
		for _, inst := range st.Instances {
			t.Logf("  instance: %s %s", inst.Flavor, inst.Root)
		}
		t.Fatalf("expected Detect to find npm install; installed=%v", st.Installed)
	}
}
