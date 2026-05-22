package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func updateHyper(source string, updater updater) (commandOutput, *hyperError) {
	url := resolveUpdateURL(source)
	stdoutText := "Updating Hyper Run from " + url + "\n"
	result, err := updater.update(url)
	if err != nil {
		return commandOutput{Stdout: stdoutText}, newError("Update failed: "+err.Error(), 1)
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
	source = firstNonBlank(strings.TrimSpace(source), strings.TrimSpace(os.Getenv("HYPER_RUN_UPDATE_SOURCE")), defaultUpdateRepo)
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return source
	}
	source = strings.TrimPrefix(source, "github:")
	return fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", source, updateAssetName())
}

func updateAssetName() string {
	name := fmt.Sprintf("hyper-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

type updater interface {
	update(url string) (updateResult, error)
}

type updateResult struct {
	Target         string
	FallbackUsed   bool
	FallbackReason string
	Warning        string
}

type realUpdater struct{}

func (realUpdater) update(url string) (updateResult, error) {
	response, err := http.Get(url)
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

	currentInstallErr := ""
	current, err := os.Executable()
	if err == nil && strings.TrimSpace(current) != "" {
		if err := installDownloadedBinary(downloadPath, current); err == nil {
			return updateResult{Target: current}, nil
		} else {
			currentInstallErr = err.Error()
		}
	}

	fallback := userInstallPath()
	if err := installDownloadedBinary(downloadPath, fallback); err != nil {
		return updateResult{}, fmt.Errorf("could not install fallback %s: %s", fallback, err.Error())
	}
	result := updateResult{Target: fallback, FallbackUsed: true, FallbackReason: currentInstallErr}
	if !pathContains(filepath.Dir(fallback)) {
		result.Warning = filepath.Dir(fallback) + " is not on PATH"
	}
	return result, nil
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
