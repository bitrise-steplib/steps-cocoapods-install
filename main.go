package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-core/bitrise-init/utility"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/rubycommand"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/cache"
)

// ConfigsModel ...
type ConfigsModel struct {
	SourceRootPath          string
	PodfilePath             string
	IsUpdateCocoapods       string
	InstallCocoapodsVersion string
	Verbose                 string
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		SourceRootPath:          os.Getenv("source_root_path"),
		PodfilePath:             os.Getenv("podfile_path"),
		IsUpdateCocoapods:       os.Getenv("is_update_cocoapods"),
		InstallCocoapodsVersion: os.Getenv("install_cocoapods_version"),
		Verbose:                 os.Getenv("verbose"),
	}
}

func (configs ConfigsModel) print() {
	log.Infof("Configs:")
	log.Printf("- SourceRootPath: %s", configs.SourceRootPath)
	log.Printf("- PodfilePath: %s", configs.PodfilePath)
	log.Printf("- IsUpdateCocoapods: %s", configs.IsUpdateCocoapods)
	log.Printf("- InstallCocoapodsVersion: %s", configs.InstallCocoapodsVersion)
	log.Printf("- Verbose: %s", configs.Verbose)
}

func (configs ConfigsModel) validate() error {
	if configs.SourceRootPath == "" {
		return errors.New("no SourceRootPath parameter specified")
	}
	if exist, err := pathutil.IsDirExists(configs.SourceRootPath); err != nil {
		return fmt.Errorf("failed to check if SourceRootPath exists at: %s, error: %s", configs.SourceRootPath, err)
	} else if !exist {
		return fmt.Errorf("SourceRootPath does not exist at: %s", configs.SourceRootPath)
	}

	if configs.PodfilePath != "" {
		if exist, err := pathutil.IsPathExists(configs.PodfilePath); err != nil {
			return fmt.Errorf("failed to check if PodfilePath exists at: %s, error: %s", configs.PodfilePath, err)
		} else if !exist {
			return fmt.Errorf("PodfilePath does not exist at: %s", configs.PodfilePath)
		}
	}

	if configs.IsUpdateCocoapods != "true" && configs.IsUpdateCocoapods != "false" {
		return fmt.Errorf("IsUpdateCocoapods invalid value: %s, available: [true false]", configs.IsUpdateCocoapods)
	}

	if configs.Verbose != "" {
		if configs.Verbose != "true" && configs.Verbose != "false" {
			return fmt.Errorf(`invalid Verbose parameter specified: %s, available: ["true", "false"]`, configs.Verbose)
		}
	}

	return nil
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func findMostRootPodfileInFileList(fileList []string) (string, error) {
	podfiles, err := utility.FilterPaths(fileList,
		utility.AllowPodfileBaseFilter,
		utility.ForbidGitDirComponentFilter,
		utility.ForbidPodsDirComponentFilter,
		utility.ForbidCarthageDirComponentFilter,
		utility.ForbidFramworkComponentWithExtensionFilter)
	if err != nil {
		return "", err
	}

	podfiles, err = utility.SortPathsByComponents(podfiles)
	if err != nil {
		return "", err
	}

	if len(podfiles) < 1 {
		return "", nil
	}

	return podfiles[0], nil
}

func findMostRootPodfile(dir string) (string, error) {
	fileList, err := utility.ListPathInDirSortedByComponents(dir, false)
	if err != nil {
		return "", err
	}

	return findMostRootPodfileInFileList(fileList)
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
	configs := createConfigsModelFromEnvs()

	fmt.Println()
	configs.print()

	if err := configs.validate(); err != nil {
		failf("Issue with input: %s", err)
	}

	if configs.IsUpdateCocoapods != "false" {
		log.Warnf("`is_update_cocoapods` is deprecated!")
		log.Warnf("CocoaPods version is determined based on the Podfile.lock or the Gemfile.lock in the Podfile's directory.")
	}

	if configs.InstallCocoapodsVersion != "" {
		log.Warnf("`install_cocoapods_version` is deprecated!")
		log.Warnf("CocoaPods version is determined based on the Podfile.lock or the Gemfile.lock in the Podfile's directory.")
	}

	fmt.Println()
	log.Printf("System installed cocoapods version:")

	podVersionCmdSlice := []string{"pod", "--version"}

	log.Donef("$ %s", command.PrintableCommandArgs(false, podVersionCmdSlice))

	cmd, err := rubycommand.New("pod", "--version")
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		failf("Command failed, error: %s", err)
	}

	//
	// Search for Podfile
	podfilePath := ""

	if configs.PodfilePath == "" {
		fmt.Println()
		log.Infof("Searching for Podfile")

		absSourceRootPath, err := pathutil.AbsPath(configs.SourceRootPath)
		if err != nil {
			failf("Failed to expand (%s), error: %s", configs.SourceRootPath, err)
		}

		absPodfilePath, err := findMostRootPodfile(absSourceRootPath)
		if err != nil {
			failf("Failed to find Podfile, error: %s", err)
		}
		if absPodfilePath == "" {
			failf("No Podfile found")
		}

		log.Donef("Found Podfile: %s", absPodfilePath)

		podfilePath = absPodfilePath
	} else {
		absPodfilePath, err := pathutil.AbsPath(configs.PodfilePath)
		if err != nil {
			failf("Failed to expand (%s), error: %s", configs.PodfilePath, err)
		}

		fmt.Println()
		log.Infof("Using Podfile: %s", absPodfilePath)

		podfilePath = absPodfilePath
	}

	podfileDir := filepath.Dir(podfilePath)

	//
	// Install required cocoapods version
	fmt.Println()
	log.Infof("Determining required cocoapods version")

	useBundler := false
	useCocoapodsVersion := ""

	log.Printf("Searching for Podfile.lock")

	// Check Podfile.lock for CocoaPods version
	podfileLockPth := filepath.Join(podfileDir, "Podfile.lock")
	if exist, err := pathutil.IsPathExists(podfileLockPth); err != nil {
		failf("Failed to check Podfile.lock at: %s, error: %s", podfileLockPth, err)
	} else if exist {
		// Podfile.lock exist scearch for version
		log.Printf("Found Podfile.lock: %s", podfileLockPth)

		version, err := cocoapodsVersionFromPodfileLock(podfileLockPth)
		if err != nil {
			failf("Failed to determine CocoaPods version, error: %s", err)
		}

		if version != "" {
			useCocoapodsVersion = version
			log.Donef("Required CocoaPods version (from Podfile.lock): %s", useCocoapodsVersion)
		} else {
			log.Warnf("No CocoaPods version found in Podfile.lock! (%s)", podfileLockPth)
		}
	} else {
		log.Warnf("No Podfile.lock found at: %s", podfileLockPth)
		log.Warnf("Make sure it's committed into your repository!")
	}

	// Collecting caches
	fmt.Println()
	log.Infof("Collecting Pod cache...")

	podsCache := cache.New()
	if absPodsDirPth, err := filepath.Abs(filepath.Join(podfileDir, "Pods")); err != nil {
		log.Warnf("Cache collection skipped: failed to determine (Pods) dir path")
	} else {
		if absPodfileLockPth, err := filepath.Abs(podfileLockPth); err != nil {
			log.Warnf("Cache collection skipped: failed to determine (Podfile.lock) path")
		} else {
			podsCache.IncludePath(fmt.Sprintf("%s -> %s", absPodsDirPth, absPodfileLockPth))

			if err := podsCache.Commit(); err != nil {
				log.Warnf("Cache collection skipped: failed to commit cache paths.")
			}
		}
	}

	if useCocoapodsVersion == "" {
		gemfileLockPth := filepath.Join(podfileDir, "Gemfile.lock")
		log.Printf("Searching for Gemfile.lock with cocoapods gem")

		if exist, err := pathutil.IsPathExists(gemfileLockPth); err != nil {
			failf("Failed to check Gemfile.lock at: %s, error: %s", gemfileLockPth, err)
		} else if exist {
			version, err := cocoapodsVersionFromGemfileLock(gemfileLockPth)
			if err != nil {
				failf("Failed to check if Gemfile.lock contains cocopods, error: %s", err)
			}

			if version != "" {
				log.Printf("Found Gemfile.lock: %s", gemfileLockPth)
				log.Donef("Gemfile.lock defined cocoapods version: %s", version)

				useBundler = true
			}
		} else {
			log.Printf("No Gemfile.lock with cocoapods gem found at: %s", gemfileLockPth)
			log.Donef("Using system installed CocoaPods version")
		}
	}

	// Install cocoapods version
	fmt.Println()
	log.Infof("Install cocoapods version")

	podCmdSlice := []string{"pod"}

	if useBundler {
		log.Printf("Install cocoapods with bundler")

		bundleInstallCmd := []string{"bundle", "install", "--jobs", "20", "--retry", "5"}

		log.Donef("$ %s", command.PrintableCommandArgs(false, bundleInstallCmd))

		cmd, err := rubycommand.NewFromSlice(bundleInstallCmd...)
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetDir(podfileDir)

		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}
	} else if useCocoapodsVersion != "" {
		log.Printf("Checking cocoapods %s gem", useCocoapodsVersion)

		installed, err := rubycommand.IsGemInstalled("cocoapods", useCocoapodsVersion)
		if err != nil {
			failf("Failed to check if cocoapods %s installed, error: %s", useCocoapodsVersion, err)
		}

		if !installed {
			log.Printf("Installing")

			cmds, err := rubycommand.GemInstall("cocoapods", useCocoapodsVersion)
			if err != nil {
				failf("Failed to create command model, error: %s", err)
			}

			for _, cmd := range cmds {
				log.Donef("$ %s", cmd.PrintableCommandArgs())

				cmd.SetDir(podfileDir)

				if err := cmd.Run(); err != nil {
					failf("Command failed, error: %s", err)
				}
			}
		} else {
			log.Printf("Installed")
		}

		podCmdSlice = append(podCmdSlice, fmt.Sprintf("_%s_", useCocoapodsVersion))
	} else {
		log.Printf("Using system installed cocoapods")
	}

	// Run pod install
	fmt.Println()
	log.Infof("Installing Pods")

	podInstallCmdSlice := append(podCmdSlice, "install", "--no-repo-update")

	if configs.Verbose == "true" {
		podInstallCmdSlice = append(podInstallCmdSlice, "--verbose")
	}

	if useBundler {
		podInstallCmdSlice = append([]string{"bundle", "exec"}, podInstallCmdSlice...)
	}

	log.Donef("$ %s", command.PrintableCommandArgs(false, podInstallCmdSlice))

	cmd, err = rubycommand.NewFromSlice(podInstallCmdSlice...)
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(podfileDir)

	if err := cmd.Run(); err != nil {
		log.Warnf("Command failed, error: %s, retrying without --no-repo-update ...", err)

		// Repo update
		podRepoUpdateCmdSlice := append(podCmdSlice, "repo", "update")

		if useBundler {
			podInstallCmdSlice = append([]string{"bundle", "exec"}, podRepoUpdateCmdSlice...)
		}

		log.Donef("$ %s", command.PrintableCommandArgs(false, podRepoUpdateCmdSlice))

		cmd, err = rubycommand.NewFromSlice(podRepoUpdateCmdSlice...)
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}

		// Pod install
		podInstallCmdSlice := append(podCmdSlice, "install")

		if configs.Verbose == "true" {
			podInstallCmdSlice = append(podInstallCmdSlice, "--verbose")
		}

		if useBundler {
			podInstallCmdSlice = append([]string{"bundle", "exec"}, podInstallCmdSlice...)
		}

		log.Donef("$ %s", command.PrintableCommandArgs(false, podInstallCmdSlice))

		cmd, err = rubycommand.NewFromSlice(podInstallCmdSlice...)
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}
	}

	log.Donef("Success!")
}
