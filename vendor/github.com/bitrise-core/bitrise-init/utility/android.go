package utility

import (
	"sort"
	"strings"
)

const (
	buildGradleBasePath = "build.gradle"
	gradlewBasePath     = "gradlew"
)

// FixedGradlewPath ...
func FixedGradlewPath(gradlewPth string) string {
	split := strings.Split(gradlewPth, "/")
	if len(split) != 1 {
		return gradlewPth
	}

	if !strings.HasPrefix(gradlewPth, "./") {
		return "./" + gradlewPth
	}
	return gradlewPth
}

// FilterRootBuildGradleFiles ...
func FilterRootBuildGradleFiles(fileList []string) ([]string, error) {
	allowBuildGradleBaseFilter := BaseFilter(buildGradleBasePath, true)
	gradleFiles, err := FilterPaths(fileList, allowBuildGradleBaseFilter)
	if err != nil {
		return []string{}, err
	}

	if len(gradleFiles) == 0 {
		return []string{}, nil
	}

	sortableFiles := []SortablePath{}
	for _, pth := range gradleFiles {
		sortable, err := NewSortablePath(pth)
		if err != nil {
			return []string{}, err
		}
		sortableFiles = append(sortableFiles, sortable)
	}

	sort.Sort(BySortablePathComponents(sortableFiles))
	mindDepth := len(sortableFiles[0].Components)

	rootGradleFiles := []string{}
	for _, sortable := range sortableFiles {
		depth := len(sortable.Components)
		if depth == mindDepth {
			rootGradleFiles = append(rootGradleFiles, sortable.Pth)
		}
	}

	return rootGradleFiles, nil
}

// FilterGradlewFiles ...
func FilterGradlewFiles(fileList []string) ([]string, error) {
	allowGradlewBaseFilter := BaseFilter(gradlewBasePath, true)
	gradlewFiles, err := FilterPaths(fileList, allowGradlewBaseFilter)
	if err != nil {
		return []string{}, err
	}

	fixedGradlewFiles := []string{}
	for _, gradlewFile := range gradlewFiles {
		fixed := FixedGradlewPath(gradlewFile)
		fixedGradlewFiles = append(fixedGradlewFiles, fixed)
	}

	return fixedGradlewFiles, nil
}
