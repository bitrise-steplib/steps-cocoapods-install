package main

import (
	"testing"

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
