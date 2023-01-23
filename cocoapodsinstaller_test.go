package main

import (
	"strings"
	"testing"

	"bitrise-steplib/steps-cocoapods-install/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GivenCocoapodsInstaller_WhenArgsGiven_ThenRunsExpectedCommand(t *testing.T) {
	type args struct {
		podArg     []string
		podCmd     string
		podfileDir string
		verbose    bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantCmd []string
	}{
		{
			name:    "simple pod install",
			args:    args{podArg: []string{"pod"}, podCmd: "install", verbose: false},
			wantErr: false,
			wantCmd: []string{"pod", "install", "--no-repo-update"},
		},
		{
			name:    "verbose pod install",
			args:    args{podArg: []string{"pod"}, podCmd: "install", verbose: true},
			wantErr: false,
			wantCmd: []string{"pod", "install", "--no-repo-update", "--verbose"},
		},
		{
			name:    "verbose pod update",
			args:    args{podArg: []string{"pod"}, podCmd: "update", verbose: true},
			wantErr: false,
			wantCmd: []string{"pod", "update", "--no-repo-update", "--verbose"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			cmd := new(mocks.Command)
			cmd.On("PrintableCommandArgs").Return(strings.Join(tt.wantCmd, " "))
			cmd.On("Run").Return(nil)

			cmdFactory := new(mocks.CommandFactory)
			cmdFactory.On("Create", tt.wantCmd[0], tt.wantCmd[1:], mock.Anything).Return(cmd)

			installer := NewCocoapodsInstaller(cmdFactory)

			// When
			err := installer.InstallPods(tt.args.podArg, tt.args.podCmd, tt.args.podfileDir, tt.args.verbose)

			// Then
			if (err != nil) != tt.wantErr {
				t.Errorf("CocoapodsInstaller.InstallPods() error = %v, wantErr %v", err, tt.wantErr)
			}
			cmdFactory.AssertExpectations(t)
			cmd.AssertExpectations(t)
		})
	}
}

func Test_GivenCocoapodsInstaller_WhenCommandFails_ThenPrintsErrors(t *testing.T) {
	expectedErrors := []string{
		"[!] Error installing boost",
		"[!] /usr/bin/curl -f -L -o /var/folders/v9/hjkgcpmn6bq99p7gvyhpq6800000gn/T/d20221018-7204-3bfs7/file.tbz https://boostorg.jfrog.io/artifactory/main/release/1.76.0/source/boost_1_76_0.tar.bz2 --create-dirs --netrc-optional --retry 2 -A 'CocoaPods/1.11.3 cocoapods-downloader/1.6.3'",
		"curl: (22) The requested URL returned error: 504 Gateway Time-out",
	}
	errorFinder := cocoapodsCmdErrorFinder{}
	errors := errorFinder.findErrors(failingPodInstallOutput)
	require.Equal(t, expectedErrors, errors)
	require.Equal(t, true, errorFinder.IsTransientProblem)
}

const failingPodInstallOutput = `Installing YogaKit (1.18.1)
Installing boost (1.76.0)

[!] Error installing boost
[!] /usr/bin/curl -f -L -o /var/folders/v9/hjkgcpmn6bq99p7gvyhpq6800000gn/T/d20221018-7204-3bfs7/file.tbz https://boostorg.jfrog.io/artifactory/main/release/1.76.0/source/boost_1_76_0.tar.bz2 --create-dirs --netrc-optional --retry 2 -A 'CocoaPods/1.11.3 cocoapods-downloader/1.6.3'

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
Warning: Transient problem: HTTP error Will retry in 1 seconds. 2 retries 
Warning: left.

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:02 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:03 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:04 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:05 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:06 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:07 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:08 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:09 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:11 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:12 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:13 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:14 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:15 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:16 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:17 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:18 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:19 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:20 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:21 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:22 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:23 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:24 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:26 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:27 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:28 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:29 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:30 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:31 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:32 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:33 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:34 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:35 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:36 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:37 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:38 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:39 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:40 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:41 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:42 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:43 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:45 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:46 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:47 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:48 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:49 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:50 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:51 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:52 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:53 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:54 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:55 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:56 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:57 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:58 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:59 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:02 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:03 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:04 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:05 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:06 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:07 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:08 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:09 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:10 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:11 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:12 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:13 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:14 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:15 --:--:--     0
Warning: Transient problem: HTTP error Will retry in 2 seconds. 1 retries 
Warning: left.

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:02 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:03 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:04 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:05 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:06 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:07 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:08 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:10 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:11 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:12 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:13 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:14 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:15 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:16 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:17 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:18 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:19 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:20 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:21 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:22 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:23 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:25 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:26 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:27 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:28 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:29 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:30 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:31 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:32 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:33 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:34 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:35 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:36 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:37 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:38 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:39 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:40 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:42 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:43 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:44 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:45 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:46 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:47 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:48 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:49 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:51 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:52 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:53 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:54 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:55 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:56 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:57 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:58 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:59 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:00 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:02 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:04 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:05 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:06 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:07 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:08 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:09 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:10 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:11 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:12 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:13 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:14 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:15 --:--:--     0
curl: (22) The requested URL returned error: 504 Gateway Time-out
`
