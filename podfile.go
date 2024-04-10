package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

const specsRepoWarning = `### CocoaPods tip
Your Podfile is still using the Specs repo. Switch to the CDN source for faster and more reliable dependency installs!
Learn more about the one-line change [here](https://blog.cocoapods.org/CocoaPods-1.8.0-beta/).
`

// isPodfileUsingSpecsRepo returns true if the Podfile contains a source 'https://github.com/CocoaPods/Specs.git'.
// It returns false if the CDN source or any other 3rd party git source is used.
func isPodfileUsingSpecsRepo(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	specsRepoDefined := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		cleanLine := strings.ReplaceAll(line, "\"", "'")
		cleanLine = strings.ToLower(cleanLine)
		if cleanLine == "source 'https://github.com/cocoapods/specs.git'" {
			specsRepoDefined = true
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return specsRepoDefined, nil
}

func addSpecsRepoAnnotation(cmdFactory command.Factory) {
	cmd := cmdFactory.Create("bitrise", []string{":annotations", "annotate", specsRepoWarning, "--style", "info"}, nil)
	_ = cmd.Run() // ignore error, this is best-effort
}
