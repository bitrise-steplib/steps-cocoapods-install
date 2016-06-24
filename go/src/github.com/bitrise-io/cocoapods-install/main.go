package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	log "github.com/bitrise-io/cocoapods-install/logger"
	"github.com/bitrise-io/cocoapods-install/run"
	"github.com/bitrise-io/cocoapods-install/sorting"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

func validateRequiredInput(key, value string) {
	if value == "" {
		log.Fail("Missing required input: %s", key)
	}
}

func isRelevantPodfile(podfile string) bool {
	if strings.Contains(podfile, ".git/") {
		return false
	}

	pathComponents := strings.Split(podfile, string(filepath.Separator))
	for _, component := range pathComponents {
		if component == "Carthage" {
			return false
		}
	}

	return true
}

func findMostRootPodfile(sourceRootPth string) (string, error) {
	// Search for Podfile in root dir
	rootPodfilePath := filepath.Join(sourceRootPth, "Podfile")
	if exist, err := pathutil.IsPathExists(rootPodfilePath); err != nil {
		return "", err
	} else if exist {
		return rootPodfilePath, nil
	}

	// Search for most root Podfile
	podfiles := []string{}
	pattern := filepath.Join(sourceRootPth, "*/Podfile")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	for _, podfile := range matches {
		if isRelevantPodfile(podfile) {
			podfiles = append(podfiles, podfile)
		}
	}

	if len(podfiles) == 0 {
		return "", nil
	}

	sort.Sort(sorting.ByComponents(podfiles))
	return podfiles[0], nil
}

func cocoapodsVersionFromGemfileLockContent(content string) string {
	relevantLines := []string{}
	lines := strings.Split(content, "\n")

	specsStart := false
	for _, line := range lines {
		if strings.Contains(line, "specs:") {
			specsStart = true
		}

		trimmed := strings.Trim(line, " ")
		if trimmed == "" {
			break
		}

		if specsStart {
			relevantLines = append(relevantLines, line)
		}
	}

	exp := regexp.MustCompile(`cocoapods \((.+)\)`)
	for _, line := range relevantLines {
		match := exp.FindStringSubmatch(line)
		if match != nil && len(match) == 2 {
			return match[1]
		}
	}

	return ""
}

func cocoapodsVersionFromGemfileLock(gemfileLockPth string) (string, error) {
	content, err := fileutil.ReadStringFromFile(gemfileLockPth)
	if err != nil {
		return "", err
	}
	return cocoapodsVersionFromGemfileLockContent(content), nil
}

func cocoapodsVersionFromPodfileLockContent(content string) string {
	exp := regexp.MustCompile("COCOAPODS: (.+)")
	match := exp.FindStringSubmatch(content)
	if match != nil && len(match) == 2 {
		return match[1]
	}
	return ""
}

func cocoapodsVersionFromPodfileLock(podfileLockPth string) (string, error) {
	content, err := fileutil.ReadStringFromFile(podfileLockPth)
	if err != nil {
		return "", err
	}
	return cocoapodsVersionFromPodfileLockContent(content), nil
}

func main() {
	//
	// Inputs
	sourceRootPath := os.Getenv("source_root_path")
	podfilePath := os.Getenv("podfile_path")

	systemCocoapodsVersion, err := run.GetPodVersion()
	if err != nil {
		log.Fail("Failed to get system installed pod version, err: %s", err)
	}

	log.Configs(sourceRootPath, podfilePath, systemCocoapodsVersion)
	validateRequiredInput("source_root_path", sourceRootPath)

	//
	// Search for Podfile
	if podfilePath == "" {
		log.Info("Searching for Podfile")
		var err error
		podfilePath, err = findMostRootPodfile(sourceRootPath)
		if err != nil {
			log.Fail("Failed to find podfiles, err: %s", err)
		}

		if podfilePath == "" {
			log.Fail("No Podfile found")
		}

		log.Done("Found Podfile: %s", podfilePath)
	}

	podfileDir := filepath.Dir(podfilePath)

	//
	// Install required cocoapods version
	log.Info("Determining required cocoapods version")

	useCocoapodsFromGemfile := false
	useCocoapodsVersion := ""

	gemfileLockPth := filepath.Join(podfileDir, "Gemfile.lock")
	log.Details("Searching for Gemfile.lock with cocoapods gem")

	if exist, err := pathutil.IsPathExists(gemfileLockPth); err != nil {
		log.Fail("Failed to check Gemfile.lock at: %s, error: %s", gemfileLockPth, err)
	} else if exist {
		version, err := cocoapodsVersionFromGemfileLock(gemfileLockPth)
		if err != nil {
			log.Fail("Failed to check if Gemfile.lock contains cocopods, error: %s", err)
		}

		if version != "" {
			log.Details("Found Gemfile.lock: %s", gemfileLockPth)
			log.Done("Gemfile.lock defined cocoapods version: %s", version)

			bundleInstallCmd := []string{"bundle", "install"}
			if err := run.CmdSlice(podfileDir, false, bundleInstallCmd); err != nil {
				log.Fail("Command failed, error: %s", err)
			}

			useCocoapodsFromGemfile = true
		}
	}

	if !useCocoapodsFromGemfile {
		log.Details("Searching for Podfile.lock")
		// Check Podfile.lock for CocoaPods version
		podfileLockPth := filepath.Join(podfileDir, "Podfile.lock")
		if exist, err := pathutil.IsPathExists(podfileLockPth); err != nil {
			log.Fail("Failed to check Podfile.lock at: %s, error: %s", podfileLockPth, err)
		} else if exist {
			// Podfile.lock exist scearch for version
			log.Details("Found Podfile.lock: %s", podfileLockPth)

			version, err := cocoapodsVersionFromPodfileLock(podfileLockPth)
			if err != nil {
				log.Fail("Failed to determin CocoaPods version, error: %s", err)
			}

			log.Done("CocoaPods version: %s", version)
			if version != systemCocoapodsVersion {
				useCocoapodsVersion = version
			}
		} else {
			// Use system installed cocoapods
			log.Done("Using system installed CocoaPods version")
		}
	}

	// Install cocoapods version
	podCmd := []string{"pod"}
	if !useCocoapodsFromGemfile && useCocoapodsVersion != "" {
		installed, err := run.CheckForGemInstalled("cocoapods", useCocoapodsVersion)
		if err != nil {
			log.Fail("Command failed, error: %s", err)
		}

		if !installed {
			log.Info("Installing cocoapods: %s", useCocoapodsVersion)
			gemInstallCocoapodsCmd := []string{"sudo", "gem", "install", "cocoapods", "-v", useCocoapodsVersion}
			if err := run.CmdSlice(podfileDir, false, gemInstallCocoapodsCmd); err != nil {
				log.Fail("Command failed, error: %s", err)
			}
		}

		podCmd = append(podCmd, fmt.Sprintf("_%s_", useCocoapodsVersion))
	}

	// Run pod install
	log.Info("Installing Pods")

	if err := run.FixCocoapodsSSHSourceInDir(podfilePath); err != nil {
		log.Fail("Failed to fix CocoaPods ssh source, err: %s", err)
	}

	podVersionCmd := append(podCmd, "--version")
	if err := run.CmdSlice(podfileDir, useCocoapodsFromGemfile, podVersionCmd); err != nil {
		log.Fail("Command failed, error: %s", err)
	}

	podInstallNoUpdateCmd := append(podCmd, "install", "--verbose", "--no-repo-update")
	if err := run.CmdSlice(podfileDir, useCocoapodsFromGemfile, podInstallNoUpdateCmd); err != nil {
		log.Warn("Command failed with error: %s, retrying without --no-repo-update ...", err)

		podRepoUpdateCmd := append(podCmd, "repo", "update")
		if err := run.CmdSlice(podfileDir, useCocoapodsFromGemfile, podRepoUpdateCmd); err != nil {
			log.Fail("Command failed, error: %s", err)
		}

		podInstallCmd := append(podCmd, "install", "--verbose")
		if err := run.CmdSlice(podfileDir, useCocoapodsFromGemfile, podInstallCmd); err != nil {
			log.Fail("Command failed, error: %s", err)
		}
	}
	log.Done("Succed")
}
