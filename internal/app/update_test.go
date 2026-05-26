package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyChecksumFilePassesForMatchingAsset(t *testing.T) {
	root := t.TempDir()
	binaryPath := filepath.Join(root, "hyper-darwin-arm64")
	checksumsPath := filepath.Join(root, "checksums.txt")
	writeFile(t, binaryPath, "hello")
	writeFile(t, checksumsPath, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  hyper-darwin-arm64\n")

	if err := verifyChecksumFile(binaryPath, checksumsPath, "hyper-darwin-arm64"); err != nil {
		t.Fatalf("checksum verification failed: %v", err)
	}
}

func TestVerifyChecksumFileRejectsMismatch(t *testing.T) {
	root := t.TempDir()
	binaryPath := filepath.Join(root, "hyper-darwin-arm64")
	checksumsPath := filepath.Join(root, "checksums.txt")
	writeFile(t, binaryPath, "hello")
	writeFile(t, checksumsPath, "0000000000000000000000000000000000000000000000000000000000000000  hyper-darwin-arm64\n")

	err := verifyChecksumFile(binaryPath, checksumsPath, "hyper-darwin-arm64")
	if err == nil {
		t.Fatal("expected checksum mismatch")
	}
	assertContains(t, err.Error(), "checksum mismatch")
}

func TestVerifyChecksumFileRequiresAssetEntry(t *testing.T) {
	root := t.TempDir()
	binaryPath := filepath.Join(root, "hyper-darwin-arm64")
	checksumsPath := filepath.Join(root, "checksums.txt")
	writeFile(t, binaryPath, "hello")
	writeFile(t, checksumsPath, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  hyper-linux-amd64\n")

	err := verifyChecksumFile(binaryPath, checksumsPath, "hyper-darwin-arm64")
	if err == nil {
		t.Fatal("expected missing checksum")
	}
	assertContains(t, err.Error(), "checksum not found")
}

func TestSignatureVerificationPlanRequiresCosignWhenRequested(t *testing.T) {
	t.Setenv("HYPER_RUN_VERIFY_SIGNATURE", "required")
	t.Setenv("PATH", t.TempDir())
	_, _, err := signatureVerificationPlan()
	if err == nil {
		t.Fatal("expected missing cosign to fail when signature verification is required")
	}
	assertContains(t, err.Error(), "requires cosign")
}

func TestSignatureVerificationPlanSkipsWithoutCosignByDefault(t *testing.T) {
	t.Setenv("HYPER_RUN_VERIFY_SIGNATURE", "")
	t.Setenv("PATH", t.TempDir())
	verify, skip, err := signatureVerificationPlan()
	if err != nil {
		t.Fatalf("signature plan failed: %v", err)
	}
	if verify {
		t.Fatal("expected signature verification to be skipped when cosign is unavailable")
	}
	assertContains(t, skip, "cosign not found")
}

func TestSignatureVerificationPlanUsesCosignWhenAvailable(t *testing.T) {
	dir := t.TempDir()
	name := "cosign"
	if os.PathSeparator == '\\' {
		name = "cosign.exe"
	}
	writeFile(t, filepath.Join(dir, name), "")
	if err := os.Chmod(filepath.Join(dir, name), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	verify, skip, err := signatureVerificationPlan()
	if err != nil {
		t.Fatalf("signature plan failed: %v", err)
	}
	if !verify || skip != "" {
		t.Fatalf("expected verification with cosign, got verify=%v skip=%q", verify, skip)
	}
}
