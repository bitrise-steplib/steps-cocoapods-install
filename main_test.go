package main

import (
	"errors"
	"testing"

	"bitrise-steplib/steps-cocoapods-install/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindMostRootPodfile(t *testing.T) {
	t.Log("single Podfile")
	{
		fileList := []string{
			"./Podfile",
		}

		podfile, err := findMostRootPodfileInFileList(fileList)
		require.NoError(t, err)
		require.Equal(t, "./Podfile", podfile)
	}

	t.Log("single Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/Podfile",
		}

		podfile, err := findMostRootPodfileInFileList(fileList)
		require.NoError(t, err)
		require.Equal(t, "/Users/bitrise/my/podfile/dir/Podfile", podfile)
	}

	t.Log("lower case Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/podfile",
		}

		podfile, err := findMostRootPodfileInFileList(fileList)
		require.NoError(t, err)
		require.Equal(t, "/Users/bitrise/my/podfile/dir/podfile", podfile)
	}

	t.Log("multi case Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/poDfile",
		}

		podfile, err := findMostRootPodfileInFileList(fileList)
		require.NoError(t, err)
		require.Equal(t, "/Users/bitrise/my/podfile/dir/poDfile", podfile)
	}

	t.Log("multiple Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/Podfile",
			"/Users/bitrise/my/dir/Podfile",
			"/Users/bitrise/dir/Podfile",
		}

		podfile, err := findMostRootPodfileInFileList(fileList)
		require.NoError(t, err)
		require.Equal(t, "/Users/bitrise/dir/Podfile", podfile)
	}

	t.Log("multiple Podfile")
	{
		fileList := []string{
			"./my/podfile/dir/Podfile",
			"./my/dir/Podfile",
			"./dir/Podfile",
			"./",
		}

		podfile, err := findMostRootPodfileInFileList(fileList)
		require.NoError(t, err)
		require.Equal(t, "./dir/Podfile", podfile)
	}
}

func TestCocoapodsVersionFromPodfileLockContent(t *testing.T) {
	t.Log("Podfile.lock cocoapods")
	{
		content := `PODS:
  - Alamofire (3.4.0)

DEPENDENCIES:
  - Alamofire (~> 3.4)

SPEC CHECKSUMS:
  Alamofire: c19a627cefd6a95f840401c49ab1f124e07f54ee

PODFILE CHECKSUM: f2a6f4eed25b89d16fc8e906af222b4e63afa6c3

COCOAPODS: 1.0.0
`

		actual := cocoapodsVersionFromPodfileLockContent(content)
		require.Equal(t, "1.0.0", actual)
	}

	t.Log("Podfile.lock without cocoapods")
	{
		content := `PODS:
	- Alamofire (3.4.0)

DEPENDENCIES:
	- Alamofire (~> 3.4)

SPEC CHECKSUMS:
	Alamofire: c19a627cefd6a95f840401c49ab1f124e07f54ee

PODFILE CHECKSUM: f2a6f4eed25b89d16fc8e906af222b4e63afa6c3
`

		actual := cocoapodsVersionFromPodfileLockContent(content)
		require.Equal(t, "", actual)
	}
}

func TestIsIncludedInGemfileLockVersionRanges(t *testing.T) {
	t.Log("Match version")
	{
		gemfileLockVersion := "1.0.0"

		isIncluded, err := isIncludedInGemfileLockVersionRanges("1.0.0", gemfileLockVersion)
		require.NoError(t, err)
		require.True(t, isIncluded)
		isExcluded, err := isIncludedInGemfileLockVersionRanges("2.0.0", gemfileLockVersion)
		require.NoError(t, err)
		require.False(t, isExcluded)
	}

	t.Log("Specify version")
	{
		gemfileLockVersion := "~> 1.0.0"

		isIncluded, err := isIncludedInGemfileLockVersionRanges("1.0.0", gemfileLockVersion)
		require.NoError(t, err)
		require.True(t, isIncluded)
		isExcluded, err := isIncludedInGemfileLockVersionRanges("2.0.0", gemfileLockVersion)
		require.NoError(t, err)
		require.False(t, isExcluded)
	}

	t.Log("Range version")
	{
		gemfileLockVersion := ">= 1.0.0, < 2.0.0"

		isIncluded, err := isIncludedInGemfileLockVersionRanges("1.0.0", gemfileLockVersion)
		require.NoError(t, err)
		require.True(t, isIncluded)
		isExcluded, err := isIncludedInGemfileLockVersionRanges("2.0.0", gemfileLockVersion)
		require.NoError(t, err)
		require.False(t, isExcluded)
	}
}

func TestRubyBuildNeedsPipe2Workaround(t *testing.T) {
	t.Log("non-Darwin host: skips the version check, workaround not needed")
	{
		cmdFactory := new(mocks.CommandFactory)

		needsWorkaround := rubyBuildNeedsPipe2Workaround("linux", cmdFactory)

		require.False(t, needsWorkaround)
		cmdFactory.AssertExpectations(t)
	}

	t.Log("macOS below 27: workaround needed")
	{
		cmd := new(mocks.Command)
		cmd.On("RunAndReturnTrimmedOutput").Return("26.6.2", nil)

		cmdFactory := new(mocks.CommandFactory)
		cmdFactory.On("Create", "sw_vers", []string{"-productVersion"}, mock.Anything).Return(cmd)

		needsWorkaround := rubyBuildNeedsPipe2Workaround("darwin", cmdFactory)

		require.True(t, needsWorkaround)
		cmdFactory.AssertExpectations(t)
		cmd.AssertExpectations(t)
	}

	t.Log("macOS 27 or newer: workaround not needed")
	{
		cmd := new(mocks.Command)
		cmd.On("RunAndReturnTrimmedOutput").Return("27.0", nil)

		cmdFactory := new(mocks.CommandFactory)
		cmdFactory.On("Create", "sw_vers", []string{"-productVersion"}, mock.Anything).Return(cmd)

		needsWorkaround := rubyBuildNeedsPipe2Workaround("darwin", cmdFactory)

		require.False(t, needsWorkaround)
		cmdFactory.AssertExpectations(t)
		cmd.AssertExpectations(t)
	}

	t.Log("sw_vers fails: fail safe, assume workaround needed")
	{
		cmd := new(mocks.Command)
		cmd.On("RunAndReturnTrimmedOutput").Return("", errors.New("command not found"))

		cmdFactory := new(mocks.CommandFactory)
		cmdFactory.On("Create", "sw_vers", []string{"-productVersion"}, mock.Anything).Return(cmd)

		needsWorkaround := rubyBuildNeedsPipe2Workaround("darwin", cmdFactory)

		require.True(t, needsWorkaround)
		cmdFactory.AssertExpectations(t)
		cmd.AssertExpectations(t)
	}

	t.Log("sw_vers returns unparsable output: fail safe, assume workaround needed")
	{
		cmd := new(mocks.Command)
		cmd.On("RunAndReturnTrimmedOutput").Return("not-a-version", nil)

		cmdFactory := new(mocks.CommandFactory)
		cmdFactory.On("Create", "sw_vers", []string{"-productVersion"}, mock.Anything).Return(cmd)

		needsWorkaround := rubyBuildNeedsPipe2Workaround("darwin", cmdFactory)

		require.True(t, needsWorkaround)
		cmdFactory.AssertExpectations(t)
		cmd.AssertExpectations(t)
	}
}
