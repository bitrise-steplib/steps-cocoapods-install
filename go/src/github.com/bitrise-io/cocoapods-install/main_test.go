package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPathContainsComponent(t *testing.T) {
	// Should filter .git folder
	t.Log("not inside .git workspace")
	{
		actual := isPathContainsComponent("/Users/bitrise/sample-apps-ios-cocoapods/CarthageSampleAppWithCocoapods.xcworkspace", gitFolderName)
		require.Equal(t, false, actual)
	}

	t.Log("not .git project - relative path")
	{
		actual := isPathContainsComponent("CarthageSampleAppWithCocoapods.xcodeproj", gitFolderName)
		require.Equal(t, false, actual)
	}

	t.Log(".git project")
	{
		actual := isPathContainsComponent("/Users/bitrise/ios-no-shared-schemes/.git/Checkouts/Result/Result.xcodeproj", gitFolderName)
		require.Equal(t, true, actual)
	}

	t.Log(".git workspace - relative path")
	{
		actual := isPathContainsComponent(".git/Checkouts/Result/Result.xcworkspace", gitFolderName)
		require.Equal(t, true, actual)
	}

	t.Log(".git workspace - relative path")
	{
		actual := isPathContainsComponent("./sub/dir/.git/Checkouts/Result/Result.xcworkspace", gitFolderName)
		require.Equal(t, true, actual)
	}

	t.Log(".git workspace - relative path")
	{
		actual := isPathContainsComponent("sub/dir/.git/Checkouts/Result/Result.xcworkspace", gitFolderName)
		require.Equal(t, true, actual)
	}

	// Should filter Pods folder
	t.Log("not pod workspace")
	{
		actual := isPathContainsComponent("/Users/bitrise/sample-apps-ios-cocoapods/PodsSampleAppWithCocoapods.xcworkspace", podsFolderName)
		require.Equal(t, false, actual)
	}

	t.Log("not pod project - relative path")
	{
		actual := isPathContainsComponent("PodsSampleAppWithCocoapods.xcodeproj", podsFolderName)
		require.Equal(t, false, actual)
	}

	t.Log("pod project")
	{
		actual := isPathContainsComponent("/Users/bitrise/sample-apps-ios-cocoapods/Pods/Pods.xcodeproj", podsFolderName)
		require.Equal(t, true, actual)
	}

	t.Log("pod workspace - relative path")
	{
		actual := isPathContainsComponent("Pods/Pods.xcworkspace", podsFolderName)
		require.Equal(t, true, actual)
	}

	t.Log("pod workspace - relative path")
	{
		actual := isPathContainsComponent("./sub/dir/Pods/Pods.xcworkspace", podsFolderName)
		require.Equal(t, true, actual)
	}

	t.Log("pod workspace - relative path")
	{
		actual := isPathContainsComponent("sub/dir/Pods/Pods.xcworkspace", podsFolderName)
		require.Equal(t, true, actual)
	}

	// Should filter Carthage folder
	t.Log("not Carthage workspace")
	{
		actual := isPathContainsComponent("/Users/bitrise/sample-apps-ios-cocoapods/CarthageSampleAppWithCocoapods.xcworkspace", carthageFolderName)
		require.Equal(t, false, actual)
	}

	t.Log("not Carthage project - relative path")
	{
		actual := isPathContainsComponent("CarthageSampleAppWithCocoapods.xcodeproj", carthageFolderName)
		require.Equal(t, false, actual)
	}

	t.Log("Carthage project")
	{
		actual := isPathContainsComponent("/Users/bitrise/ios-no-shared-schemes/Carthage/Checkouts/Result/Result.xcodeproj", carthageFolderName)
		require.Equal(t, true, actual)
	}

	t.Log("Carthage workspace - relative path")
	{
		actual := isPathContainsComponent("Carthage/Checkouts/Result/Result.xcworkspace", carthageFolderName)
		require.Equal(t, true, actual)
	}

	t.Log("Carthage workspace - relative path")
	{
		actual := isPathContainsComponent("./sub/dir/Carthage/Checkouts/Result/Result.xcworkspace", carthageFolderName)
		require.Equal(t, true, actual)
	}

	t.Log("Carthage workspace - relative path")
	{
		actual := isPathContainsComponent("sub/dir/Carthage/Checkouts/Result/Result.xcworkspace", carthageFolderName)
		require.Equal(t, true, actual)
	}
}

func TestIsPathContainsComponentWithExtension(t *testing.T) {
	// Should filter .framework folder
	t.Log("not .framework workspace")
	{
		actual := isPathContainsComponentWithExtension("/Users/bitrise/sample-apps-ios-cocoapods/CarthageSampleAppWithCocoapods.xcworkspace", frameworkExt)
		require.Equal(t, false, actual)
	}

	t.Log("not .framework project - relative path")
	{
		actual := isPathContainsComponentWithExtension("CarthageSampleAppWithCocoapods.xcodeproj", frameworkExt)
		require.Equal(t, false, actual)
	}

	t.Log(".framework project")
	{
		actual := isPathContainsComponentWithExtension("/Users/bitrise/ios-no-shared-schemes/test.framework/Checkouts/Result/Result.xcodeproj", frameworkExt)
		require.Equal(t, true, actual)
	}

	t.Log(".framework workspace - relative path")
	{
		actual := isPathContainsComponentWithExtension("test.framework/Checkouts/Result/Result.xcworkspace", frameworkExt)
		require.Equal(t, true, actual)
	}

	t.Log(".framework workspace - relative path")
	{
		actual := isPathContainsComponentWithExtension("./sub/dir/test.framework/Checkouts/Result/Result.xcworkspace", frameworkExt)
		require.Equal(t, true, actual)
	}

	t.Log(".framework workspace - relative path")
	{
		actual := isPathContainsComponentWithExtension("sub/dir/test.framework/Checkouts/Result/Result.xcworkspace", frameworkExt)
		require.Equal(t, true, actual)
	}
}

func TestIsRelevantPodfile(t *testing.T) {
	t.Log(`.git, pod, carthage, .framework - not relevant`)
	{
		fileList := []string{
			"/Users/bitrise/.git/Podfile",
			"/Users/bitrise/sample-apps-ios-cocoapods/Pods/Podfile",
			"/Users/bitrise/ios-no-shared-schemes/Carthage/Checkouts/Result/Podfile",
			"/Users/bitrise/ios-no-shared-schemes/test.framework/Checkouts/Result/Podfile",
		}

		for _, file := range fileList {
			require.Equal(t, false, isRelevantPodfile(file))
		}
	}

	require.Equal(t, true, isRelevantPodfile("Podfile"))
	require.Equal(t, true, isRelevantPodfile("/Podfile"))

	require.Equal(t, false, isRelevantPodfile("Carthage/Podfile"))
	require.Equal(t, false, isRelevantPodfile(".git/Podfile"))

	require.Equal(t, false, isRelevantPodfile("Podfile.lock"))
	require.Equal(t, false, isRelevantPodfile("Gemfile"))
}

func TestFindMostRootPodfile(t *testing.T) {
	t.Log("single Podfile")
	{
		fileList := []string{
			"./Podfile",
		}

		podfile := findMostRootPodfile(fileList)
		require.Equal(t, "./Podfile", podfile)
	}

	t.Log("single Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/Podfile",
		}

		podfile := findMostRootPodfile(fileList)
		require.Equal(t, "/Users/bitrise/my/podfile/dir/Podfile", podfile)
	}

	t.Log("lower case Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/podfile",
		}

		podfile := findMostRootPodfile(fileList)
		require.Equal(t, "/Users/bitrise/my/podfile/dir/podfile", podfile)
	}

	t.Log("multi case Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/poDfile",
		}

		podfile := findMostRootPodfile(fileList)
		require.Equal(t, "/Users/bitrise/my/podfile/dir/poDfile", podfile)
	}

	t.Log("multiple Podfile")
	{
		fileList := []string{
			"/Users/bitrise/my/podfile/dir/Podfile",
			"/Users/bitrise/my/dir/Podfile",
			"/Users/bitrise/dir/Podfile",
		}

		podfile := findMostRootPodfile(fileList)
		require.Equal(t, "/Users/bitrise/dir/Podfile", podfile)
	}

	t.Log("multiple Podfile")
	{
		fileList := []string{
			"./my/podfile/dir/Podfile",
			"./my/dir/Podfile",
			"./dir/Podfile",
		}

		podfile := findMostRootPodfile(fileList)
		require.Equal(t, "./dir/Podfile", podfile)
	}
}

func TestCocoapodsVersionFromGemfileLockContent(t *testing.T) {
	t.Log("Gemfile.lock with cocoapods")
	{
		content := `GEM
  remote: https://rubygems.org/
  specs:
    activesupport (4.2.6)
      i18n (~> 0.7)
      json (~> 1.7, >= 1.7.7)
      minitest (~> 5.1)
      thread_safe (~> 0.3, >= 0.3.4)
      tzinfo (~> 1.1)
    claide (1.0.0)
    cocoapods (1.0.0)
      activesupport (>= 4.0.2)
      claide (>= 1.0.0, < 2.0)
      cocoapods-core (= 1.0.0)
      cocoapods-deintegrate (>= 1.0.0, < 2.0)
      cocoapods-downloader (>= 1.0.0, < 2.0)
      cocoapods-plugins (>= 1.0.0, < 2.0)
      cocoapods-search (>= 1.0.0, < 2.0)
      cocoapods-stats (>= 1.0.0, < 2.0)
      cocoapods-trunk (>= 1.0.0, < 2.0)
      cocoapods-try (>= 1.0.0, < 2.0)
      colored (~> 1.2)
      escape (~> 0.0.4)
      fourflusher (~> 0.3.0)
      molinillo (~> 0.4.5)
      nap (~> 1.0)
      xcodeproj (>= 1.0.0, < 2.0)
    cocoapods-core (1.0.0)
      activesupport (>= 4.0.2)
      fuzzy_match (~> 2.0.4)
      nap (~> 1.0)
    cocoapods-deintegrate (1.0.0)
    cocoapods-downloader (1.0.0)
    cocoapods-plugins (1.0.0)
      nap
    cocoapods-search (1.0.0)
    cocoapods-stats (1.0.0)
    cocoapods-trunk (1.0.0)
      nap (>= 0.8, < 2.0)
      netrc (= 0.7.8)
    cocoapods-try (1.0.0)
    colored (1.2)
    escape (0.0.4)
    fourflusher (0.3.0)
    fuzzy_match (2.0.4)
    i18n (0.7.0)
    json (1.8.3)
    minitest (5.9.0)
    molinillo (0.4.5)
    nap (1.1.0)
    netrc (0.7.8)
    thread_safe (0.3.5)
    tzinfo (1.2.2)
      thread_safe (~> 0.1)
    xcodeproj (1.0.0)
      activesupport (>= 3)
      claide (>= 1.0.0, < 2.0)
      colored (~> 1.2)

PLATFORMS
  ruby

DEPENDENCIES
  cocoapods (~> 1.0)

BUNDLED WITH
   1.10.6
  `

		actual := cocoapodsVersionFromGemfileLockContent(content)
		require.Equal(t, "1.0.0", actual)
	}

	t.Log("Gemfile.lock without cocoapods")
	{
		content := `GEM
remote: https://rubygems.org/
specs:
  activesupport (4.2.6)
    i18n (~> 0.7)
    json (~> 1.7, >= 1.7.7)
    minitest (~> 5.1)
    thread_safe (~> 0.3, >= 0.3.4)
    tzinfo (~> 1.1)
  claide (1.0.0)
  cocoapods-core (1.0.0)
    activesupport (>= 4.0.2)
    fuzzy_match (~> 2.0.4)
    nap (~> 1.0)
  cocoapods-deintegrate (1.0.0)
  cocoapods-downloader (1.0.0)
  cocoapods-plugins (1.0.0)
    nap
  cocoapods-search (1.0.0)
  cocoapods-stats (1.0.0)
  cocoapods-trunk (1.0.0)
    nap (>= 0.8, < 2.0)
    netrc (= 0.7.8)
  cocoapods-try (1.0.0)
  colored (1.2)
  escape (0.0.4)
  fourflusher (0.3.0)
  fuzzy_match (2.0.4)
  i18n (0.7.0)
  json (1.8.3)
  minitest (5.9.0)
  molinillo (0.4.5)
  nap (1.1.0)
  netrc (0.7.8)
  thread_safe (0.3.5)
  tzinfo (1.2.2)
    thread_safe (~> 0.1)
  xcodeproj (1.0.0)
    activesupport (>= 3)
    claide (>= 1.0.0, < 2.0)
    colored (~> 1.2)

PLATFORMS
ruby

DEPENDENCIES
cocoapods (~> 1.0)

BUNDLED WITH
 1.10.6
`

		actual := cocoapodsVersionFromGemfileLockContent(content)
		require.Equal(t, "", actual)
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
