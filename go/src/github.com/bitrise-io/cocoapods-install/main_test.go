package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRelevantPodfile(t *testing.T) {
	require.Equal(t, true, isRelevantPodfile("Podfile"))
	require.Equal(t, true, isRelevantPodfile("/Podfile"))

	require.Equal(t, false, isRelevantPodfile("Carthage/Podfile"))
	require.Equal(t, false, isRelevantPodfile(".git/Podfile"))
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
