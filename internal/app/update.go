package app

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func updateHyper(source string, updater updater) (commandOutput, *hyperError) {
	request := resolveUpdateRequest(source)
	stdoutText := "Updating Hyper Run from " + request.DownloadURL + "\n"
	result, err := updater.update(request)
	if err != nil {
		return commandOutput{Stdout: stdoutText}, newError("Update failed: "+err.Error(), 1)
	}
	if result.ChecksumVerified {
		stdoutText += "Verified checksum: " + result.ChecksumAsset + "\n"
	}
	if result.Target != "" {
		stdoutText += "Installed executable: " + result.Target + "\n"
	}
	if result.FallbackUsed {
		stdoutText += "Current executable could not be replaced; installed to the user bin fallback.\n"
	}
	if result.FallbackReason != "" {
		stdoutText += "Fallback reason: " + result.FallbackReason + "\n"
	}
	if result.Warning != "" {
		stdoutText += "Warning: " + result.Warning + "\n"
	}
	stdoutText += "Run `hyper version` to verify the active executable.\n"
	stdoutText += "Hyper Run update completed.\n"
	return stdout(stdoutText), nil
}

func resolveUpdateURL(source string) string {
	return resolveUpdateRequest(source).DownloadURL
}

func resolveUpdateRequest(source string) updateRequest {
	source = firstNonBlank(strings.TrimSpace(source), strings.TrimSpace(os.Getenv("HYPER_RUN_UPDATE_SOURCE")), defaultUpdateRepo)
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		asset := assetNameFromDownloadURL(source)
		return updateRequest{
			DownloadURL: source,
			ChecksumURL: strings.TrimSpace(os.Getenv("HYPER_RUN_CHECKSUM_URL")),
			AssetName:   asset,
		}
	}
	source = strings.TrimPrefix(source, "github:")
	asset := updateAssetName()
	return updateRequest{
		DownloadURL: fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", source, asset),
		ChecksumURL: fmt.Sprintf("https://github.com/%s/releases/latest/download/checksums.txt", source),
		AssetName:   asset,
	}
}

func assetNameFromDownloadURL(downloadURL string) string {
	withoutQuery := strings.SplitN(downloadURL, "?", 2)[0]
	withoutFragment := strings.SplitN(withoutQuery, "#", 2)[0]
	return filepath.Base(strings.TrimRight(withoutFragment, "/"))
}

func updateAssetName() string {
	name := fmt.Sprintf("hyper-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

type updater interface {
	update(request updateRequest) (updateResult, error)
}

type updateRequest struct {
	DownloadURL string
	ChecksumURL string
	AssetName   string
}

type updateResult struct {
	Target           string
	FallbackUsed     bool
	FallbackReason   string
	Warning          string
	ChecksumVerified bool
	ChecksumAsset    string
}

type realUpdater struct{}

func (realUpdater) update(request updateRequest) (updateResult, error) {
	response, err := http.Get(request.DownloadURL)
	if err != nil {
		return updateResult{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return updateResult{}, fmt.Errorf("download returned %s", response.Status)
	}
	downloadPath := filepath.Join(os.TempDir(), fmt.Sprintf("hyper-download-%d", os.Getpid()))
	file, err := os.OpenFile(downloadPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return updateResult{}, err
	}
	if _, err := io.Copy(file, response.Body); err != nil {
		file.Close()
		return updateResult{}, err
	}
	if err := file.Close(); err != nil {
		return updateResult{}, err
	}
	defer os.Remove(downloadPath)
	if err := os.Chmod(downloadPath, 0755); err != nil {
		return updateResult{}, err
	}
	checksumVerified := false
	if request.ChecksumURL != "" {
		if request.AssetName == "" {
			return updateResult{}, fmt.Errorf("checksum verification requires an asset name")
		}
		if err := verifyRemoteChecksum(downloadPath, request.ChecksumURL, request.AssetName); err != nil {
			return updateResult{}, err
		}
		checksumVerified = true
	}

	currentInstallErr := ""
	current, err := os.Executable()
	if err == nil && strings.TrimSpace(current) != "" {
		if err := installDownloadedBinary(downloadPath, current); err == nil {
			return updateResult{Target: current, ChecksumVerified: checksumVerified, ChecksumAsset: request.AssetName}, nil
		} else {
			currentInstallErr = err.Error()
		}
	}

	fallback := userInstallPath()
	if err := installDownloadedBinary(downloadPath, fallback); err != nil {
		return updateResult{}, fmt.Errorf("could not install fallback %s: %s", fallback, err.Error())
	}
	result := updateResult{Target: fallback, FallbackUsed: true, FallbackReason: currentInstallErr, ChecksumVerified: checksumVerified, ChecksumAsset: request.AssetName}
	if !pathContains(filepath.Dir(fallback)) {
		result.Warning = filepath.Dir(fallback) + " is not on PATH"
	}
	return result, nil
}

func verifyRemoteChecksum(path, checksumURL, asset string) error {
	checksumsPath := filepath.Join(os.TempDir(), fmt.Sprintf("hyper-checksums-%d", os.Getpid()))
	defer os.Remove(checksumsPath)
	response, err := http.Get(checksumURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("checksum download returned %s", response.Status)
	}
	file, err := os.OpenFile(checksumsPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, response.Body); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return verifyChecksumFile(path, checksumsPath, asset)
}

func verifyChecksumFile(path, checksumsPath, asset string) error {
	expected, err := checksumForAsset(checksumsPath, asset)
	if err != nil {
		return err
	}
	actual, err := sha256File(path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", asset, expected, actual)
	}
	return nil
}

func checksumForAsset(checksumsPath, asset string) (string, error) {
	file, err := os.Open(checksumsPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == asset {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("checksum not found for %s in checksums.txt", asset)
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

func installDownloadedBinary(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	temp := filepath.Join(filepath.Dir(target), fmt.Sprintf(".hyper-update-%d", os.Getpid()))
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(temp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		os.Remove(temp)
		return err
	}
	if err := output.Close(); err != nil {
		os.Remove(temp)
		return err
	}
	if err := os.Chmod(temp, 0755); err != nil {
		os.Remove(temp)
		return err
	}
	return os.Rename(temp, target)
}

func userInstallPath() string {
	if override := strings.TrimSpace(os.Getenv("HYPER_INSTALL_PATH")); override != "" {
		return override
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return filepath.Join(".", "hyper")
	}
	name := "hyper"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(home, ".local", "bin", name)
}

func pathContains(dir string) bool {
	for _, part := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.Clean(part) == filepath.Clean(dir) {
			return true
		}
	}
	return false
}
