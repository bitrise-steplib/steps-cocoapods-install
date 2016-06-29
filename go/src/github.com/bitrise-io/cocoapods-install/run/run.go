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
	systemRubyPth = "/usr/bin/ruby"
	brewRubyPth   = "/usr/local/bin/ruby"
)

// ----------------------
// RubyCommand

// RubyInstallType ...
type RubyInstallType int8

const (
	// SystemRuby ...
	SystemRuby RubyInstallType = iota
	// BrewRuby ...
	BrewRuby
	// RVMRuby ...
	RVMRuby
	// RbenvRuby ...
	RbenvRuby
)

// RubyCommandModel ...
type RubyCommandModel struct {
	rubyInstallType RubyInstallType
}

// NewRubyCommandModel ...
func NewRubyCommandModel() (RubyCommandModel, error) {
	whichRuby, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("which", "ruby")
	if err != nil {
		return RubyCommandModel{}, err
	}

	command := RubyCommandModel{}

	if whichRuby == systemRubyPth {
		command.rubyInstallType = SystemRuby
	} else if whichRuby == brewRubyPth {
		command.rubyInstallType = BrewRuby
	} else if cmdExist([]string{"rvm", "-v"}) {
		command.rubyInstallType = RVMRuby
	} else if cmdExist([]string{"rbenv", "-v"}) {
		command.rubyInstallType = RbenvRuby
	} else {
		return RubyCommandModel{}, errors.New("unkown ruby installation type")
	}

	return command, nil
}

// Execute ...
func (command RubyCommandModel) Execute(workDir string, useBundle bool, cmdSlice []string) error {
	if useBundle {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	if command.rubyInstallType == SystemRuby {
		cmdSlice = append([]string{"sudo"}, cmdSlice...)
	}

	return execute(workDir, false, cmdSlice)
}

// ExecuteForOutput ...
func (command RubyCommandModel) ExecuteForOutput(workDir string, useBundle bool, cmdSlice []string) (string, error) {
	if useBundle {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	if command.rubyInstallType == SystemRuby {
		cmdSlice = append([]string{"sudo"}, cmdSlice...)
	}

	return executeForOutput(workDir, false, cmdSlice)
}

// GemInstall ...
func (command RubyCommandModel) GemInstall(gem, version string) error {
	cmdSlice := []string{"gem", "install", gem, "-v", version, "--no-document"}
	if err := command.Execute("", false, cmdSlice); err != nil {
		return err
	}

	if command.rubyInstallType == RbenvRuby {
		cmdSlice := []string{"rbenv", "rehash"}

		if err := command.Execute("", false, cmdSlice); err != nil {
			return err
		}
	}

	return nil
}

// IsGemInstalled ...
func (command RubyCommandModel) IsGemInstalled(gem, version string) (bool, error) {
	cmdSlice := []string{"gem", "list"}
	out, err := command.ExecuteForOutput("", false, cmdSlice)
	if err != nil {
		return false, err
	}

	regexpStr := gem + ` \((?P<versions>.*)\)`
	exp := regexp.MustCompile(regexpStr)
	matches := exp.FindStringSubmatch(out)
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

// GetPodVersion ...
func (command RubyCommandModel) GetPodVersion() string {
	cmdSlice := []string{"pod", "--version"}
	out, err := command.ExecuteForOutput("", false, cmdSlice)
	if err != nil {
		return ""
	}

	return out
}

// ----------------------
// Common

func execute(workDir string, bundleExec bool, cmdSlice []string) error {
	if len(cmdSlice) == 0 {
		return errors.New("no command specified")
	}

	if bundleExec {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	prinatableCmd := cmdex.PrintableCommandArgs(false, cmdSlice)
	log.Details("=> %s", prinatableCmd)

	if len(cmdSlice) == 1 {
		out, err := cmdex.RunCommandInDirAndReturnCombinedStdoutAndStderr(workDir, cmdSlice[0])
		log.Details(out)

		return err
	}

	out, err := cmdex.RunCommandInDirAndReturnCombinedStdoutAndStderr(workDir, cmdSlice[0], cmdSlice[1:len(cmdSlice)]...)
	log.Details(out)

	return err
}

func executeForOutput(workDir string, bundleExec bool, cmdSlice []string) (string, error) {
	if len(cmdSlice) == 0 {
		return "", errors.New("no command specified")
	}

	if bundleExec {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	if len(cmdSlice) == 1 {
		return cmdex.RunCommandInDirAndReturnCombinedStdoutAndStderr(workDir, cmdSlice[0])
	}

	return cmdex.RunCommandInDirAndReturnCombinedStdoutAndStderr(workDir, cmdSlice[0], cmdSlice[1:len(cmdSlice)]...)
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
