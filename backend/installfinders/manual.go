package installfinders

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/settings"
)

// manualInstallName makes manual installs tell each other apart in the install picker,
// since they all share the same launcher name
func manualInstallName(path string) string {
	return fmt.Sprintf("Manual (%s)", filepath.Base(path))
}

// FindManualInstallations turns the paths listed in the settings file into installations,
// using exactly the same detection as the launcher based finders
// Anything that fails to be detected is reported as an error instead,
// so that it shows up as an invalid install, like a broken launcher install would
func FindManualInstallations() ([]*common.Installation, []error) {
	manualInstalls := settings.Settings.ManualInstalls

	installs := make([]*common.Installation, 0, len(manualInstalls))
	findErrors := make([]error, 0)

	for _, manualInstall := range manualInstalls {
		installation, err := MakeManualInstallation(manualInstall.Path, manualInstall.Branch)
		if err != nil {
			findErrors = append(findErrors, common.InstallFindError{
				Path:  manualInstall.Path,
				Inner: err,
			})
			continue
		}
		installs = append(installs, installation)
	}

	return installs, findErrors
}

// MakeManualInstallation detects what game install is at path and builds the installation for it
func MakeManualInstallation(path string, branch string) (*common.Installation, error) {
	branchValue, err := parseGameBranch(branch)
	if err != nil {
		return nil, err
	}

	platform := common.NativePlatform()

	installType, version, savedPath, err := common.GetGameInfo(path, platform)
	if err != nil {
		return nil, fmt.Errorf("no Satisfactory installation found at this location: %w", err)
	}

	installation := &common.Installation{
		Path:      filepath.Clean(path),
		Version:   version,
		Type:      installType,
		Location:  common.LocationTypeLocal,
		Branch:    branchValue,
		Launcher:  manualInstallName(path),
		SavedPath: savedPath,
	}

	// Only client installs can be started, and only on the platform they run on
	if installType == common.InstallTypeWindowsClient && platform.Os() == "windows" {
		executablePath := common.GetClientExecutablePath(path)
		if executablePath == "" {
			slog.Warn("manual install has no executable to start the game with", slog.String("path", path))
		} else {
			installation.LaunchPath = []string{executablePath}
		}
	}

	return installation, nil
}

// parseGameBranch reads the branch from the settings file, which is typed by hand,
// so it does not insist on the exact casing the game uses
func parseGameBranch(branch string) (common.GameBranch, error) {
	switch strings.ToLower(branch) {
	case "":
		return common.BranchStable, nil
	case strings.ToLower(string(common.BranchStable)):
		return common.BranchStable, nil
	case strings.ToLower(string(common.BranchExperimental)):
		return common.BranchExperimental, nil
	default:
		return "", fmt.Errorf("unknown game branch %q, use %q or %q", branch, common.BranchStable, common.BranchExperimental)
	}
}
