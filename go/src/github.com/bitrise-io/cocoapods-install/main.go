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

func fileList(searchDir string) ([]string, error) {
	fileList := []string{}

	if err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	}); err != nil {
		return []string{}, err
	}

	return fileList, nil
}

func isRelevantPodfile(pth string) bool {
	basename := filepath.Base(pth)
	if basename != "Podfile" {
		return false
	}

	if strings.Contains(pth, ".git/") {
		return false
	}

	pathComponents := strings.Split(pth, string(filepath.Separator))
	for _, component := range pathComponents {
		if component == "Carthage" {
			return false
		}
	}

	return true
}

func findMostRootPodfile(fileList []string) (string, error) {
	podfiles := []string{}

	for _, file := range fileList {
		if isRelevantPodfile(file) {
			podfiles = append(podfiles, file)
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
	if os.Getenv("is_update_cocoapods") != "false" {
		log.Warn("`is_update_cocoapods` is deprecated!")
		log.Warn("CocoaPods version is determined based on the Gemfile.lock or the Podfile.lock in the Podfile's directory.")
	}

	if os.Getenv("install_cocoapods_version") != "" {
		log.Warn("`install_cocoapods_version` is deprecated!")
		log.Warn("CocoaPods version is determined based on the Gemfile.lock or the Podfile.lock in the Podfile's directory.")
	}

	sourceRootPath := os.Getenv("source_root_path")
	podfilePath := os.Getenv("podfile_path")

	rubyCommand, err := run.NewRubyCommandModel()
	if err != nil {
		log.Fail("Failed to create ruby command, err: %s", err)
	}

	systemCocoapodsVersion := rubyCommand.GetPodVersion()

	log.Configs(sourceRootPath, podfilePath, systemCocoapodsVersion)
	validateRequiredInput("source_root_path", sourceRootPath)

	//
	// Search for Podfile
	if podfilePath == "" {
		log.Info("Searching for Podfile")

		fileList, err := fileList(sourceRootPath)
		if err != nil {
			log.Fail("Failed to list files at: %s", sourceRootPath)
		}

		podfilePath, err = findMostRootPodfile(fileList)
		if err != nil {
			log.Fail("Failed to find root Podfile, err: %s", err)
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
			if err := rubyCommand.Execute(podfileDir, false, bundleInstallCmd); err != nil {
				log.Fail("Command failed, error: %s", err)
			}

			useCocoapodsFromGemfile = true
		}
	} else {
		log.Details("No Gemfile.lock with cocoapods gem found at: %s", gemfileLockPth)
	}

	if !useCocoapodsFromGemfile {
		fmt.Println("")
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
			log.Warn("No Podfile.lock found at: %s", podfileLockPth)
			log.Warn("Make sure it's committed into your repository!")
			log.Done("Using system installed CocoaPods version")
		}
	}

	// Install cocoapods version
	podCmd := []string{"pod"}
	if !useCocoapodsFromGemfile && useCocoapodsVersion != "" {
		installed, err := rubyCommand.IsGemInstalled("cocoapods", useCocoapodsVersion)
		if err != nil {
			log.Fail("Command failed, error: %s", err)
		}

		if !installed {
			log.Info("Installing cocoapods: %s", useCocoapodsVersion)
			if err := rubyCommand.GemInstall("cocoapods", useCocoapodsVersion); err != nil {
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
	if err := rubyCommand.Execute(podfileDir, useCocoapodsFromGemfile, podVersionCmd); err != nil {
		log.Fail("Command failed, error: %s", err)
	}

	podInstallNoUpdateCmd := append(podCmd, "install", "--verbose", "--no-repo-update")
	if err := rubyCommand.Execute(podfileDir, useCocoapodsFromGemfile, podInstallNoUpdateCmd); err != nil {
		log.Warn("Command failed with error: %s, retrying without --no-repo-update ...", err)

		podRepoUpdateCmd := append(podCmd, "repo", "update")
		if err := rubyCommand.Execute(podfileDir, useCocoapodsFromGemfile, podRepoUpdateCmd); err != nil {
			log.Fail("Command failed, error: %s", err)
		}

		podInstallCmd := append(podCmd, "install", "--verbose")
		if err := rubyCommand.Execute(podfileDir, useCocoapodsFromGemfile, podInstallCmd); err != nil {
			log.Fail("Command failed, error: %s", err)
		}
	}
	log.Done("Success!")
}
