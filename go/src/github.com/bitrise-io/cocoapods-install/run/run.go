package run

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/bitrise-io/cocoapods-install/logger"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

const (
	osxSystemRubyPth = "/usr/bin/ruby"
	brewRubyPth      = "/usr/local/bin/ruby"
)

// CmdSlice ...
func CmdSlice(workDir string, bundleExec bool, cmdSlice []string) error {
	if bundleExec {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	prinatableCmd := cmdex.PrintableCommandArgs(false, cmdSlice)
	log.Details("=> %s", prinatableCmd)

	out, err := cmdex.RunCommandInDirAndReturnCombinedStdoutAndStderr(workDir, cmdSlice[0], cmdSlice[1:len(cmdSlice)]...)
	log.Details(out)

	return err
}

// GetPodVersion ...
func GetPodVersion() (string, error) {
	if installed, err := CheckForGemInstalled("cocoapods", ""); err != nil {
		return "", err
	} else if !installed {
		return "", nil
	}

	cmdSlice := []string{"pod", "--version"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(cmdSlice[0], cmdSlice[1:len(cmdSlice)]...)
	if err != nil {
		return "", err
	}

	split := strings.Split(out, "\n")
	return split[len(split)-1], nil
}

// FixCocoapodsSSHSourceInDir ...
func FixCocoapodsSSHSourceInDir(podfilePth string) error {
	podRepoFixCounter := 0

	applySourceFix := func(URIStr string) error {
		podRepoFixCounter++

		repoURLString := URIStr
		repoAliasName := fmt.Sprintf("SourceFix-%d", podRepoFixCounter)

		fixCmd := []string{"pod", "repo", "add", repoAliasName, repoURLString}

		// remove previously applied fix - if this fix script
		//  would be called multiple times
		homeDir := pathutil.UserHomeDir()
		repoAliasPth := filepath.Join(homeDir, ".cocoapods/repos", repoAliasName)
		out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("rm", "-rf", repoAliasPth)
		if err != nil {
			return fmt.Errorf("error: %s, out: %s", err, out)
		}

		// apply fix
		out, err = cmdex.RunCommandAndReturnCombinedStdoutAndStderr(fixCmd[0], fixCmd[1:len(fixCmd)]...)
		if err != nil {
			return fmt.Errorf("error: %s, out: %s", err, out)
		}

		return nil
	}

	absPodfilePth, err := pathutil.AbsPath(podfilePth)
	if err != nil {
		return err
	}

	content, err := fileutil.ReadStringFromFile(absPodfilePth)
	if err != nil {
		return err
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		lineStrinp := strings.Trim(line, " ")
		parts := strings.Split(lineStrinp, " ")
		if len(parts) >= 2 && strings.ToLower(parts[0]) == "source" {
			expectedURIPart := strings.Trim(line, `"`)
			expectedURIPart = strings.Trim(expectedURIPart, `'`)

			url, err := url.Parse(expectedURIPart)
			if err != nil {
				if err := applySourceFix(expectedURIPart); err != nil {
					return err
				}
			}

			if url.Scheme == "ssh" {
				if err := applySourceFix(expectedURIPart); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// CheckForGemInstalled ...
func CheckForGemInstalled(gem, version string) (bool, error) {
	cmdSlice := []string{"gem", "list"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(cmdSlice[0], cmdSlice[1:len(cmdSlice)]...)
	if err != nil {
		return false, err
	}

	cocoapodsExp := regexp.MustCompile(`cocoapods \((?P<versions>.*)\)`)
	matches := cocoapodsExp.FindStringSubmatch(out)
	if len(matches) > 1 {
		if version == "" {
			return true, nil
		}

		versionsStr := matches[1]
		versions := strings.Split(versionsStr, ", ")

		for _, v := range versions {
			if v == version {
				return true, nil
			}
		}
	}

	return false, nil
}

func cmdExist(cmdSlice []string) bool {
	if len(cmdSlice) == 0 {
		return false
	}

	if len(cmdSlice) == 1 {
		_, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(cmdSlice[0])
		return (err == nil)
	}

	_, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr(cmdSlice[0], cmdSlice[1:]...)
	return (err == nil)
}

// GemInstall ...
func GemInstall(gem, version string) error {
	whichRuby, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("which", "ruby")
	if err != nil {
		return err
	}

	if whichRuby == osxSystemRubyPth {
		log.Details("using system ruby - requires sudo")

		gemInstallCocoapodsCmd := []string{"sudo", "gem", "install", "cocoapods", "-v", version, "--no-document"}
		if err := CmdSlice("", false, gemInstallCocoapodsCmd); err != nil {
			return err
		}
	} else if whichRuby == brewRubyPth {
		log.Details("using brew %s ruby", brewRubyPth)

		gemInstallCocoapodsCmd := []string{"gem", "install", "cocoapods", "-v", version, "--no-document"}
		if err := CmdSlice("", false, gemInstallCocoapodsCmd); err != nil {
			return err
		}
	} else if cmdExist([]string{"rvm", "-v"}) {
		log.Details("installing with RVM")

		gemInstallCocoapodsCmd := []string{"gem", "install", "cocoapods", "-v", version, "--no-document"}
		if err := CmdSlice("", false, gemInstallCocoapodsCmd); err != nil {
			return err
		}
	} else if cmdExist([]string{"rbenv", "-v"}) {
		log.Details("installing with rbenv")

		gemInstallCocoapodsCmd := []string{"gem", "install", "cocoapods", "-v", version, "--no-document"}
		if err := CmdSlice("", false, gemInstallCocoapodsCmd); err != nil {
			return err
		}

		rbenvRehashCmd := []string{"rbenv", "rehash"}
		if err := CmdSlice("", false, rbenvRehashCmd); err != nil {
			return err
		}
	} else {
		return errors.New("no ruby is available")
	}

	return nil
}
