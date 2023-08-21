package main

import (
	"bufio"
	"os"
	"strings"
)

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
		if strings.ReplaceAll(line, "\"", "'") == "source 'https://github.com/CocoaPods/Specs.git'" {
			specsRepoDefined = true
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return specsRepoDefined, nil
}
