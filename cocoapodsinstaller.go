package main

import (
	"os"

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
	cmdSlice := podInstallCmdSlice(podArg, podCmd, verbose)
	cmd := createPodCommand(i.rubyCmdFactory, cmdSlice, podfileDir)
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	err := cmd.Run()
	if err == nil {
		return nil
	}

	log.Warnf("pod install failed: %s, retrying with repo update...", err)

	cmdSlice = podRepoUpdateCmdSlice(podArg, verbose)
	cmd = createPodCommand(i.rubyCmdFactory, cmdSlice, podfileDir)
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return err
	}

	cmdSlice = podInstallCmdSlice(podArg, podCmd, verbose)
	cmd = createPodCommand(i.rubyCmdFactory, cmdSlice, podfileDir)
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
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

func createPodCommand(factory ruby.CommandFactory, args []string, dir string) command.Command {
	return factory.Create(args[0], args[1:], &command.Opts{
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       nil,
		Env:         nil,
		Dir:         dir,
		ErrorFinder: cocoapodsCmdErrorFinder,
	})
}

func cocoapodsCmdErrorFinder(out string) []string {
	return nil
}
