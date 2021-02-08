package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitrise-io/bitrise-init/scanners/ios"
	"github.com/bitrise-io/bitrise-init/utility"
	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/gems"
	"github.com/bitrise-io/go-utils/command/rubycommand"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/pkg/errors"
)

// ConfigsModel ...
type ConfigsModel struct {
	SourceRootPath  string
	PodfilePath     string
	Verbose         string
	IsCacheDisabled string
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		SourceRootPath:  os.Getenv("source_root_path"),
		PodfilePath:     os.Getenv("podfile_path"),
		Verbose:         os.Getenv("verbose"),
		IsCacheDisabled: os.Getenv("is_cache_disabled"),
	}
}

func (configs ConfigsModel) print() {
	log.Infof("Configs:")
	log.Printf("- SourceRootPath: %s", configs.SourceRootPath)
	log.Printf("- PodfilePath: %s", configs.PodfilePath)
	log.Printf("- Verbose: %s", configs.Verbose)
	log.Printf("- IsCacheDisabled: %s", configs.IsCacheDisabled)
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
		ios.AllowPodfileBaseFilter,
		ios.ForbidCarthageDirComponentFilter,
		ios.ForbidPodsDirComponentFilter,
		ios.ForbidGitDirComponentFilter,
		ios.ForbidFramworkComponentWithExtensionFilter)
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

func cocoapodsVersionFromPodfileLockContent(content string) string {
	exp := regexp.MustCompile("COCOAPODS: (.+)")
	match := exp.FindStringSubmatch(content)
	if len(match) == 2 {
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

// VersionSpec ...
type VersionSpec struct {
	Operator string
	Version  string
}

func splitOperatorAndVersion(input string) (VersionSpec, error) {
	splittedString := strings.Split(input, " ")
	cnt := len(splittedString)

	if cnt == 1 {
		out := VersionSpec{"", splittedString[0]}
		return out, nil
	}

	if cnt != 2 {
		err := fmt.Errorf("Invalid version range: %s", input)
		return VersionSpec{}, err
	}

	out := VersionSpec{splittedString[0], splittedString[1]}
	return out, nil
}

func isIncludedInGemfileLockVersionRanges(input string, gemfileLockVersion string) (bool, error) {
	var splittedVersions = strings.Split(gemfileLockVersion, ", ")

	for _, each := range splittedVersions {
		versionSpec, err := splitOperatorAndVersion(each)
		if err != nil {
			return false, err
		}

		switch versionSpec.Operator {
		case "":
			if input != versionSpec.Version {
				return false, nil
			}

			continue
		case "~>":
			if input != versionSpec.Version {
				return false, nil
			}

			continue
		case ">=":
			versions := strings.Split(versionSpec.Version, ".")
			inputVersions := strings.Split(input, ".")

			for i, version := range versions {
				v1, err := strconv.Atoi(version)
				if err != nil {
					return false, err
				}

				v2, err := strconv.Atoi(inputVersions[i])
				if err != nil {
					return false, err
				}

				if i != len(versions)-1 && v1 == v2 {
					continue
				}
				if v2 >= v1 {
					break
				} else {
					return false, nil
				}
			}

			continue
		case "<":
			versions := strings.Split(versionSpec.Version, ".")
			inputVersions := strings.Split(input, ".")

			for i, version := range versions {
				v1, err := strconv.Atoi(version)
				if err != nil {
					return false, err
				}

				v2, err := strconv.Atoi(inputVersions[i])
				if err != nil {
					return false, err
				}

				if i != len(versions)-1 && v1 == v2 {
					continue
				}
				if v2 < v1 {
					break
				} else {
					return false, nil
				}
			}

			continue
		default:
			err := fmt.Errorf("Unknown version operator: %s", each)
			return false, err
		}
	}

	return true, nil
}

func main() {
	configs := createConfigsModelFromEnvs()

	fmt.Println()
	configs.print()

	if err := configs.validate(); err != nil {
		failf("Issue with input: %s", err)
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
	useCocoapodsVersionFromPodfileLock := ""
	useCocoapodsVersionFromGemfileLock := ""

	log.Printf("Searching for Podfile.lock")

	// Check Podfile.lock for CocoaPods version
	podfileLockPth := filepath.Join(podfileDir, "Podfile.lock")
	isPodfileLockExists, err := pathutil.IsPathExists(podfileLockPth)
	if err != nil {
		failf("Failed to check Podfile.lock at: %s, error: %s", podfileLockPth, err)
	}

	if isPodfileLockExists {
		// Podfile.lock exist search for version
		log.Printf("Found Podfile.lock: %s", podfileLockPth)

		version, err := cocoapodsVersionFromPodfileLock(podfileLockPth)
		if err != nil {
			failf("Failed to determine CocoaPods version, error: %s", err)
		}

		if version != "" {
			useCocoapodsVersionFromPodfileLock = version
			log.Donef("Required CocoaPods version (from Podfile.lock): %s", useCocoapodsVersionFromPodfileLock)
		} else {
			log.Warnf("No CocoaPods version found in Podfile.lock! (%s)", podfileLockPth)
		}
	} else {
		log.Warnf("No Podfile.lock found at: %s", podfileLockPth)
		log.Warnf("Make sure it's committed into your repository!")
	}

	var pod gems.Version
	var bundler gems.Version

	log.Printf("Searching for gem lockfile with cocoapods gem")

	// Check gem lockfile for CocoaPods version
	gemfileLockPth, err := gems.GemFileLockPth(podfileDir)
	if err != nil && err != gems.ErrGemLockNotFound {
		failf("Failed to check gem lockfile at: %s, error: %s", podfileDir, err)
	}

	if gemfileLockPth != "" {
		// CocoaPods exist search for version in gem lockfile
		log.Printf("Found gem lockfile: %s", gemfileLockPth)

		content, err := fileutil.ReadStringFromFile(gemfileLockPth)
		if err != nil {
			failf("failed to read file (%s) contents, error: %s", gemfileLockPth, err)
		}

		pod, err = gems.ParseVersionFromBundle("cocoapods", content)
		if err != nil {
			failf("Failed to check if gem lockfile contains cocoapods, error: %s", err)
		}

		bundler, err = gems.ParseBundlerVersion(content)
		if err != nil {
			failf("Failed to parse bundler version form cocoapods, error: %s", err)
		}

		if pod.Found {
			useCocoapodsVersionFromGemfileLock = pod.Version
			log.Donef("Required CocoaPods version (from gem lockfile): %s", useCocoapodsVersionFromGemfileLock)

			isIncludedVersionRange, err := isIncludedInGemfileLockVersionRanges(useCocoapodsVersionFromPodfileLock, useCocoapodsVersionFromGemfileLock)
			if err != nil {
				failf("Failed to compare version range in gem lockfile, error: %s", err)
			}

			if !isIncludedVersionRange {
				log.Warnf("Cocoapods version required in Podfile.lock (%s) does not match Gemfile.lock (%s). Will install Cocoapods using bundler.", useCocoapodsVersionFromPodfileLock, useCocoapodsVersionFromGemfileLock)
			}
			useBundler = true
		}
	} else {
		log.Printf("No gem lockfile with cocoapods gem found at: %s", gemfileLockPth)
		log.Donef("Using system installed CocoaPods version")
	}

	// Check ruby version
	// Run this logic only in CI environment when the ruby was installed via rbenv for the virtual machine
	if os.Getenv("CI") == "true" && rubycommand.RubyInstallType() == rubycommand.RbenvRuby {
		fmt.Println()
		log.Infof("Check selected Ruby is installed")

		rubyInstalled, rversion, err := rubycommand.IsSpecifiedRbenvRubyInstalled(configs.SourceRootPath)
		if err != nil {
			log.Errorf("Failed to check if selected ruby is installed, error: %s", err)
		}

		if !rubyInstalled {
			log.Errorf("Ruby %s is not installed", rversion)
			fmt.Println()

			cmd := command.New("rbenv", "install", rversion).SetStdout(os.Stdout).SetStderr(os.Stderr)
			log.Donef("$ %s", cmd.PrintableCommandArgs())
			if err := cmd.Run(); err != nil {
				log.Errorf("Failed to install Ruby version %s, error: %s", rversion, err)
			}
		} else {
			log.Donef("Ruby %s is installed", rversion)
		}

	}

	// Install cocoapods
	fmt.Println()
	log.Infof("Installing cocoapods")

	podCmdSlice := []string{"pod"}

	if useBundler {
		fmt.Println()
		log.Infof("Installing bundler")

		// install bundler with `gem install bundler [-v version]`
		// in some configurations, the command "bunder _1.2.3_" can return 'Command not found', installing bundler solves this
		installBundlerCommand := gems.InstallBundlerCommand(bundler)
		installBundlerCommand.SetStdout(os.Stdout).SetStderr(os.Stderr)
		installBundlerCommand.SetDir(podfileDir)

		log.Donef("$ %s", installBundlerCommand.PrintableCommandArgs())
		fmt.Println()

		if err := installBundlerCommand.Run(); err != nil {
			failf("command failed, error: %s", err)
		}

		// install gem lockfile gems with `bundle [_version_] install ...`
		fmt.Println()
		log.Infof("Installing cocoapods with bundler")

		cmd, err := gems.BundleInstallCommand(bundler)
		if err != nil {
			failf("failed to create bundle command model, error: %s", err)
		}
		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()

		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}

		if useBundler {
			podCmdSlice = append(gems.BundleExecPrefix(bundler), podCmdSlice...)
		}
	} else if useCocoapodsVersionFromPodfileLock != "" {
		log.Printf("Checking cocoapods %s gem", useCocoapodsVersionFromPodfileLock)

		installed, err := rubycommand.IsGemInstalled("cocoapods", useCocoapodsVersionFromPodfileLock)
		if err != nil {
			failf("Failed to check if cocoapods %s installed, error: %s", useCocoapodsVersionFromPodfileLock, err)
		}

		if !installed {
			log.Printf("Installing")

			cmds, err := rubycommand.GemInstall("cocoapods", useCocoapodsVersionFromPodfileLock)
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

		podCmdSlice = append(podCmdSlice, fmt.Sprintf("_%s_", useCocoapodsVersionFromPodfileLock))
	} else {
		log.Printf("Using system installed cocoapods")
	}

	fmt.Println()
	log.Infof("cocoapods version:")

	// pod can be in the PATH as an rbenv shim and pod --version will return "rbenv: pod: command not found"
	cmd, err := rubycommand.NewFromSlice(append(podCmdSlice, "--version"))
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(podfileDir)

	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		failf("command failed, error: %s", err)
	}

	// Run pod install
	fmt.Println()
	log.Infof("Installing Pods")

	podInstallCmdSlice := append(podCmdSlice, "install", "--no-repo-update")
	if configs.Verbose == "true" {
		podInstallCmdSlice = append(podInstallCmdSlice, "--verbose")
	}

	cmd, err = rubycommand.NewFromSlice(podInstallCmdSlice)
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(podfileDir)

	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		log.Warnf("Command failed, error: %s, retrying without --no-repo-update ...", err)

		// Repo update
		cmd, err = rubycommand.NewFromSlice(append(podCmdSlice, "repo", "update"))
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}

		// Pod install
		podInstallCmdSlice := append(podCmdSlice, "install")
		if configs.Verbose == "true" {
			podInstallCmdSlice = append(podInstallCmdSlice, "--verbose")
		}

		cmd, err = rubycommand.NewFromSlice(podInstallCmdSlice)
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}
	}

	// Collecting caches
	if configs.IsCacheDisabled != "true" && isPodfileLockExists {
		fmt.Println()
		log.Infof("Collecting Pod cache paths...")

		podsCache := cache.New()
		podsCache.IncludePath(fmt.Sprintf("%s -> %s", filepath.Join(podfileDir, "Pods"), podfileLockPth))

		if err := podsCache.Commit(); err != nil {
			log.Warnf("Cache collection skipped: failed to commit cache paths.")
		}
	}

	log.Donef("Success!")
}
