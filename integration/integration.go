package integration

import (
	"flasher/core"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	BinaryName = "HawkTuahFlasher"
)

type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
	PlatformMacOS   Platform = "darwin"
)

type Integration struct {
	Platform    Platform
	InstallPath string
	BinaryPath  string
}

func CurrentPlatform() (Platform, error) {
	switch runtime.GOOS {
		case "linux":
			return PlatformLinux, nil

		case "windows":
			return PlatformWindows, nil

		case "darwin":
			return PlatformMacOS, nil

		default:
			return "", core.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func New(binaryPath string) (*Integration, error) {
	platform, err := CurrentPlatform()
	if err != nil {
		return nil, err
	}

	if binaryPath == "" {
		binaryPath, err = os.Executable()
		if err != nil {
			return nil, core.Errorf("resolve executable: %w", err)
		}
	}

	binaryPath, err = filepath.Abs(binaryPath)
	if err != nil {
		return nil, core.Errorf("resolve executable path: %w", err)
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		return nil, core.Errorf("stat executable: %w", err)
	}

	if info.IsDir() {
		return nil, core.Errorf("executable path is a directory: %s", binaryPath)
	}

	return &Integration{
		Platform:    platform,
		BinaryPath:  binaryPath,
		InstallPath: defaultInstallPath(platform),
	}, nil
}

func defaultInstallPath(platform Platform) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch platform {
		case PlatformLinux:
			return filepath.Join(home, ".local", "bin")

		case PlatformMacOS:
			return filepath.Join(home, "bin")

		case PlatformWindows:
			if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
				return filepath.Join(localAppData, "HawkTuahFlasher", "bin")
			}

			return filepath.Join(home, "AppData", "Local", "HawkTuahFlasher", "bin")
	}

	return ""
}

func (integration *Integration) Install() error {
	if integration == nil {
		return core.Errorf("integration is nil")
	}

	if integration.BinaryPath == "" {
		return core.Errorf("binary path is empty")
	}

	if integration.InstallPath == "" {
		return core.Errorf("install path is empty")
	}

	if err := os.MkdirAll(integration.InstallPath, 0755); err != nil {
		return core.Errorf("create install directory: %w", err)
	}

	destination := filepath.Join(integration.InstallPath, BinaryName)

	if integration.Platform == PlatformWindows {
		destination += ".exe"
	}

	if err := copyFile(integration.BinaryPath, destination); err != nil {
		return core.Errorf("install binary: %w", err)
	}

	if integration.Platform != PlatformWindows {
		if err := os.Chmod(destination, 0755); err != nil {
			return core.Errorf("make binary executable: %w", err)
		}
	}

	switch integration.Platform {
		case PlatformLinux, PlatformMacOS:
			if err := integration.configureUnixPath(); err != nil {
				return err
			}

		case PlatformWindows:
			if err := integration.configureWindowsPath(); err != nil {
				return err
			}
	}

	return nil
}

func (integration *Integration) Uninstall() error {
	if integration == nil {
		return core.Errorf("integration is nil")
	}

	if integration.InstallPath == "" {
		return core.Errorf("install path is empty")
	}

	name := BinaryName

	if integration.Platform == PlatformWindows {
		name += ".exe"
	}

	path := filepath.Join(integration.InstallPath, name)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return core.Errorf("remove binary: %w", err)
	}

	return nil
}

func (integration *Integration) IsInstalled() bool {
	if integration == nil || integration.InstallPath == "" {
		return false
	}

	name := BinaryName

	if integration.Platform == PlatformWindows {
		name += ".exe"
	}

	path := filepath.Join(integration.InstallPath, name)

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func (integration *Integration) ExecutablePath() string {
	if integration == nil || integration.InstallPath == "" {
		return ""
	}

	name := BinaryName

	if integration.Platform == PlatformWindows {
		name += ".exe"
	}

	return filepath.Join(integration.InstallPath, name)
}

func (integration *Integration) configureUnixPath() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return core.Errorf("resolve home directory: %w", err)
	}

	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		shell = "/bin/sh"
	}

	shellName := filepath.Base(shell)

	var profile string

	switch shellName {
		case "bash":
			if integration.Platform == PlatformMacOS {
				profile = filepath.Join(home, ".bash_profile")
			} else {
				profile = filepath.Join(home, ".bashrc")
			}

		case "zsh":
			profile = filepath.Join(home, ".zshrc")

		case "fish":
			profile = filepath.Join(home, ".config", "fish", "config.fish")

		default:
			profile = filepath.Join(home, ".profile")
	}

	switch shellName {
		case "fish":
			return appendProfileLine(profile, core.Sprintf("fish_add_path %s", shellQuote(integration.InstallPath)))

		default:
			return appendProfileLine(profile, core.Sprintf("export PATH=\"%s:$PATH\"", escapeShellDoubleQuote(integration.InstallPath)))
	}
}

func (integration *Integration) configureWindowsPath() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	output, err := exec.Command("powershell.exe", "-NoProfile", "-Command", "[Environment]::GetEnvironmentVariable('Path', 'User')").Output()

	if err != nil {
		return core.Errorf("read user PATH: %w", err)
	}

	userPath := strings.TrimSpace(string(output))

	for _, entry := range strings.Split(userPath, string(os.PathListSeparator)) {
		if strings.EqualFold(filepath.Clean(entry), filepath.Clean(integration.InstallPath)) {
			return nil
		}
	}

	if userPath != "" {
		userPath += string(os.PathListSeparator)
	}

	userPath += integration.InstallPath

	_, err = exec.Command("powershell.exe", "-NoProfile", "-Command", "[Environment]::SetEnvironmentVariable('Path', $args[0], 'User')", userPath).Output()

	if err != nil {
		return core.Errorf("update user PATH: %w", err)
	}

	return nil
}

func appendProfileLine(path string, line string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return core.Errorf("read shell profile: %w", err)
	}

	content := string(data)

	if strings.Contains(content, line) {
		return nil
	}

	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += "\n# HawkTuahFlasher\n"
	content += line
	content += "\n"

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return core.Errorf("create profile directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return core.Errorf("write shell profile: %w", err)
	}

	return nil
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)

	if err != nil {
		return err
	}

	defer input.Close()

	info, err := input.Stat()
	if err != nil {
		return err
	}

	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()

	if copyErr != nil {
		return copyErr
	}

	return closeErr
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func escapeShellDoubleQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, `$`, `\$`)
	value = strings.ReplaceAll(value, "`", "\\`")

	return value
}
