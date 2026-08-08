package paths

import (
	"path/filepath"
	"testing"
)

func TestProjectsProtected(t *testing.T) {
	proj := ProjectsDir()
	if !IsProtectedPath(proj) {
		t.Fatalf("projects dir should be protected: %s", proj)
	}
	if !IsProtectedPath(filepath.Join(proj, "some-id", "project.json")) {
		t.Fatal("project file should be protected")
	}
	inst := InstallDir()
	if IsProtectedPath(inst) {
		// only if misconfigured equal — normally install is under LocalAppData\Programs
		t.Logf("install dir flagged protected (unusual): %s", inst)
	}
	if err := AssertSafeToReplaceInstallDir(); err != nil {
		t.Fatalf("install dir should be safe to replace: %v", err)
	}
}
