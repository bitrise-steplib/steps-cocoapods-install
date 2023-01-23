package main

import (
	"bufio"
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
	/*
		example 1:

		[!] Error installing boost
		[!] /usr/bin/curl -f -L -o /var/folders/v9/hjkgcpmn6bq99p7gvyhpq6800000gn/T/d20221018-3650-smj60t/file.tbz https://boostorg.jfrog.io/artifactory/main/release/1.76.0/source/boost_1_76_0.tar.bz2 --create-dirs --netrc-optional --retry 2 -A 'CocoaPods/1.11.3 cocoapods-downloader/1.6.3'
		Warning: Transient problem: HTTP error Will retry in 1 seconds. 2 retries
		Warning: left.
		...
		curl: (22) The requested URL returned error: 502 Bad Gateway

		example 2:

		[!] Error installing boost
		[!] /usr/bin/curl -f -L -o /var/folders/v9/hjkgcpmn6bq99p7gvyhpq6800000gn/T/d20221018-7204-3bfs7/file.tbz https://boostorg.jfrog.io/artifactory/main/release/1.76.0/source/boost_1_76_0.tar.bz2 --create-dirs --netrc-optional --retry 2 -A 'CocoaPods/1.11.3 cocoapods-downloader/1.6.3'
		Warning: Transient problem: HTTP error Will retry in 1 seconds. 2 retries
		Warning: left.
		...
		curl: (22) The requested URL returned error: 504 Gateway Time-out
	*/

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
	}
	if err := scanner.Err(); err != nil {
		// todo: error handling
		return nil
	}

	return errors
}
