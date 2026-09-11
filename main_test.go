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

// newSwVersFactory returns a CommandFactory mock that only answers "sw_vers -productVersion"
func newSwVersFactory(t *testing.T, output string, err error) *mocks.CommandFactory {
	t.Helper()

	cmd := new(mocks.Command)
	cmd.On("RunAndReturnTrimmedOutput").Return(output, err)

	cmdFactory := new(mocks.CommandFactory)
	cmdFactory.On("Create", "sw_vers", []string{"-productVersion"}, mock.Anything).Return(cmd)
	t.Cleanup(func() { cmd.AssertExpectations(t) })

	return cmdFactory
}

func TestRubyBuildNeedsPipe2Workaround(t *testing.T) {
	tests := []struct {
		name         string
		goos         string
		swVersOutput string
		swVersErr    error
		want         bool
	}{
		{name: "non-Darwin host skips the version check entirely", goos: "linux", want: false},
		{name: "macOS below 27", goos: "darwin", swVersOutput: "26.6.2", want: true},
		{name: "macOS 27 or newer", goos: "darwin", swVersOutput: "27.0", want: false},
		{name: "sw_vers fails: fail safe", goos: "darwin", swVersErr: errors.New("command not found"), want: true},
		{name: "sw_vers returns unparsable output: fail safe", goos: "darwin", swVersOutput: "not-a-version", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdFactory := new(mocks.CommandFactory)
			if tt.goos == "darwin" {
				cmdFactory = newSwVersFactory(t, tt.swVersOutput, tt.swVersErr)
			}

			got := rubyBuildNeedsPipe2Workaround(tt.goos, cmdFactory)

			require.Equal(t, tt.want, got)
			cmdFactory.AssertExpectations(t)
		})
	}
}

func TestBuildRubyInstallOpts(t *testing.T) {
	tests := []struct {
		name                   string
		goos                   string
		existingRubyConfigOpts string
		swVersOutput           string
		wantEnv                []string
	}{
		{
			name: "workaround not needed: RUBY_CONFIGURE_OPTS left untouched",
			goos: "linux",
		},
		{
			name:         "workaround needed, no pre-existing RUBY_CONFIGURE_OPTS: sets the pipe2/dup3 overrides",
			goos:         "darwin",
			swVersOutput: "26.6.2",
			wantEnv:      []string{"RUBY_CONFIGURE_OPTS=ac_cv_func_pipe2=no ac_cv_func_dup3=no"},
		},
		{
			name:                   "workaround needed, pre-existing RUBY_CONFIGURE_OPTS: appends rather than overwrites",
			goos:                   "darwin",
			existingRubyConfigOpts: "--with-openssl-dir=/opt/openssl",
			swVersOutput:           "26.6.2",
			wantEnv:                []string{"RUBY_CONFIGURE_OPTS=--with-openssl-dir=/opt/openssl ac_cv_func_pipe2=no ac_cv_func_dup3=no"},
		},
		{
			name:                   "workaround not needed on macOS 27+, pre-existing RUBY_CONFIGURE_OPTS: still propagated explicitly",
			goos:                   "darwin",
			existingRubyConfigOpts: "--with-openssl-dir=/opt/openssl",
			swVersOutput:           "27.0",
			wantEnv:                []string{"RUBY_CONFIGURE_OPTS=--with-openssl-dir=/opt/openssl"},
		},
		{
			name:                   "workaround not needed, non-Darwin, pre-existing RUBY_CONFIGURE_OPTS: still propagated explicitly",
			goos:                   "linux",
			existingRubyConfigOpts: "--with-openssl-dir=/opt/openssl",
			wantEnv:                []string{"RUBY_CONFIGURE_OPTS=--with-openssl-dir=/opt/openssl"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RUBY_CONFIGURE_OPTS", tt.existingRubyConfigOpts)

			cmdFactory := new(mocks.CommandFactory)
			if tt.goos == "darwin" {
				cmdFactory = newSwVersFactory(t, tt.swVersOutput, nil)
			}

			opts := buildRubyInstallOpts(tt.goos, cmdFactory)

			require.Equal(t, tt.wantEnv, opts.Env)
			cmdFactory.AssertExpectations(t)
		})
	}
}
