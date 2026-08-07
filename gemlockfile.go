package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

// gemVersion holds a gem version parsed from a Gemfile.lock, on a best effort basis.
type gemVersion struct {
	Version string
	Found   bool
}

// gemFileLockNames are the possible gem lock file names, checked in order.
var gemFileLockNames = []string{"Gemfile.lock", "gems.locked"}

// errGemLockNotFound is returned by gemFileLockPath when no gem lock file exists in the given directory.
var errGemLockNotFound = errors.New("gem lock file not found")

// gemFileLockPath returns the path of the gem lock file in searchDir, if any.
func gemFileLockPath(pathChecker pathutil.PathChecker, searchDir string) (string, error) {
	for _, gemFileName := range gemFileLockNames {
		pth := filepath.Join(searchDir, gemFileName)
		exists, err := pathChecker.IsPathExists(pth)
		if err != nil {
			return "", err
		}
		if exists {
			return pth, nil
		}
	}
	return "", errGemLockNotFound
}

// readStringFromFile reads the full content of the file at path as a string.
func readStringFromFile(fileManager fileutil.FileManager, path string) (string, error) {
	file, err := fileManager.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// parseVersionFromBundle returns the specified gem version parsed from a Gemfile.lock on a best effort basis, for logging purposes only.
//
// for "fastlane" and the following Gemfile.lock example, it returns: ">= 2.0)"
//
//	specs:
//	  CFPropertyList (3.0.0)
//	  addressable (2.6.0)
//	    public_suffix (>= 2.0.2, < 4.0)
//	  atomos (0.1.3)
//	  babosa (1.0.2)
//	  badge (0.8.5)
//	    curb (~> 0.9)
//	    fastimage (>= 1.6)
//	    fastlane (>= 2.0)
//	    mini_magick (>= 4.5)
//	  claide (1.0.2)
func parseVersionFromBundle(gemName string, gemfileLockContent string) (gemVersion, error) {
	var relevantLines []string
	lines := strings.Split(gemfileLockContent, "\n")

	specsStart := false
	for _, line := range lines {
		if strings.Trim(line, " ") == "" {
			specsStart = false
		}

		if strings.Contains(line, "specs:") {
			specsStart = true
			continue
		}

		if specsStart {
			relevantLines = append(relevantLines, line)
		}
	}

	//     fastlane (1.109.0)
	exp := regexp.MustCompile(fmt.Sprintf(`^%s \((.+)\)`, regexp.QuoteMeta(gemName)))
	for _, line := range relevantLines {
		match := exp.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			continue
		}
		if len(match) != 2 {
			return gemVersion{}, fmt.Errorf("unexpected regexp match: %v", match)
		}
		return gemVersion{Version: match[1], Found: true}, nil
	}

	return gemVersion{}, nil
}

// parseBundlerVersion returns the bundler version used to create the bundle.
func parseBundlerVersion(gemfileLockContent string) (gemVersion, error) {
	/*
		BUNDLED WITH
			1.17.1
	*/
	bundlerRegexp := regexp.MustCompile(`(?m)^BUNDLED WITH\n\s+(\S+)`)
	match := bundlerRegexp.FindStringSubmatch(gemfileLockContent)
	if match == nil {
		return gemVersion{}, nil
	}
	if len(match) != 2 {
		return gemVersion{}, fmt.Errorf("unexpected regexp match: %v", match)
	}

	return gemVersion{Version: match[1], Found: true}, nil
}

// bundleExecPrefix returns a command prefix that runs a command through bundler: "bundle [_version_] exec".
func bundleExecPrefix(bundlerVersion string) []string {
	prefix := []string{"bundle"}
	if bundlerVersion != "" {
		prefix = append(prefix, fmt.Sprintf("_%s_", bundlerVersion))
	}
	return append(prefix, "exec")
}
