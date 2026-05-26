package app

import (
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
