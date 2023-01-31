package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/ruby"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/v2/command"
)

// CocoapodsInstaller ...
type CocoapodsInstaller struct {
	rubyCmdFactory ruby.CommandFactory
}

// NewCocoapodsInstaller ...
func NewCocoapodsInstaller(rubyCmdFactory ruby.CommandFactory) CocoapodsInstaller {
	return CocoapodsInstaller{rubyCmdFactory: rubyCmdFactory}
}

// InstallPods ...
func (i CocoapodsInstaller) InstallPods(podArg []string, podCmd string, podfileDir string, verbose bool) error {
	if err := i.runPodInstall(podArg, podCmd, podfileDir, verbose); err == nil {
		return nil
	} else {
		log.Warnf("pod install failed: %s, retrying with repo update...", err)
	}

	if err := i.runPodRepoUpdate(podArg, podfileDir, verbose); err != nil {
		return err
	}

	if err := i.runPodInstall(podArg, podCmd, podfileDir, verbose); err != nil {
		return err
	}

	return nil
}

func (i CocoapodsInstaller) runPodInstall(podArg []string, podCmd string, podfileDir string, verbose bool) error {
	errorFinder := &cocoapodsCmdErrorFinder{}
	cmdSlice := podInstallCmdSlice(podArg, podCmd, verbose)
	cmd := createPodCommand(i.rubyCmdFactory, cmdSlice, podfileDir, errorFinder)
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	err := cmd.Run()
	if errorFinder.IsTransientProblem {
		return fmt.Errorf("transitent error: %w", err)
	}
	return err
}

func (i CocoapodsInstaller) runPodRepoUpdate(podArg []string, podfileDir string, verbose bool) error {
	errorFinder := &cocoapodsCmdErrorFinder{}
	cmdSlice := podRepoUpdateCmdSlice(podArg, verbose)
	cmd := createPodCommand(i.rubyCmdFactory, cmdSlice, podfileDir, errorFinder)
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	err := cmd.Run()
	if errorFinder.IsTransientProblem {
		return fmt.Errorf("transitent error: %w", err)
	}
	return err
}

func podInstallCmdSlice(podArg []string, podCmd string, verbose bool) []string {
	cmdSlice := append(podArg, podCmd, "--no-repo-update")
	if verbose {
		cmdSlice = append(cmdSlice, "--verbose")
	}
	return cmdSlice
}

func podRepoUpdateCmdSlice(podArg []string, verbose bool) []string {
	cmdSlice := append(podArg, "repo", "update")
	if verbose {
		cmdSlice = append(cmdSlice, "--verbose")
	}
	return cmdSlice
}

func createPodCommand(factory ruby.CommandFactory, args []string, dir string, errorFinder *cocoapodsCmdErrorFinder) command.Command {
	return factory.Create(args[0], args[1:], &command.Opts{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       nil,
		Env:         nil,
		Dir:         dir,
		ErrorFinder: errorFinder.findErrors,
	})
}

type cocoapodsCmdErrorFinder struct {
	IsTransientProblem bool
}

func (f *cocoapodsCmdErrorFinder) findErrors(out string) []string {
	var errors []string

	reader := strings.NewReader(out)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[!] ") ||
			strings.HasPrefix(line, "curl: ") {
			errors = append(errors, line)
		}

		if strings.HasPrefix(line, "Warning: Transient problem: ") {
			f.IsTransientProblem = true
		}
	}
	if err := scanner.Err(); err != nil {
		// todo: error handling
		return nil
	}

	return errors
}
