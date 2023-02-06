package main

import (
	"errors"
	"strings"
	"testing"

	"bitrise-steplib/steps-cocoapods-install/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GivenCocoapodsInstaller_WhenArgsGiven_ThenRunsExpectedCommand(t *testing.T) {
	type args struct {
		podArg  []string
		podCmd  string
		verbose bool
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

			logger := new(mocks.Logger)
			logger.On("Donef", mock.Anything, mock.Anything)

			installer := NewCocoapodsInstaller(cmdFactory, logger)

			// When
			err := installer.InstallPods(tt.args.podArg, tt.args.podCmd, "", tt.args.verbose)

			// Then
			if (err != nil) != tt.wantErr {
				t.Errorf("CocoapodsInstaller.InstallPods() error = %v, wantErr %v", err, tt.wantErr)
			}
			cmdFactory.AssertExpectations(t)
			cmd.AssertExpectations(t)
		})
	}
}

func Test_GivenCocoapodsInstaller_WhenInstallFails_ThenRunsRepoUpdateAndRetries(t *testing.T) {
	// Given
	podArg := []string{"pod"}
	podCmd := "install"

	firstInstallCmd := new(mocks.Command)
	firstInstallCmd.On("PrintableCommandArgs").Return(mock.Anything)
	firstInstallCmd.On("Run").Return(errors.New("[!] Error installing boost")).Once()

	repoUpdateCmd := new(mocks.Command)
	repoUpdateCmd.On("PrintableCommandArgs").Return(mock.Anything)
	repoUpdateCmd.On("Run").Return(nil).Once()

	secondInstallCmd := new(mocks.Command)
	secondInstallCmd.On("PrintableCommandArgs").Return(mock.Anything)
	secondInstallCmd.On("Run").Return(nil).Once()

	cmdFactory := new(mocks.CommandFactory)
	cmdFactory.On("Create", podArg[0], []string{podCmd, "--no-repo-update"}, mock.Anything).Return(firstInstallCmd).Once()
	cmdFactory.On("Create", podArg[0], []string{"repo", "update"}, mock.Anything).Return(repoUpdateCmd).Once()
	cmdFactory.On("Create", podArg[0], []string{podCmd, "--no-repo-update"}, mock.Anything).Return(secondInstallCmd).Once()

	logger := new(mocks.Logger)
	logger.On("Donef", mock.Anything, mock.Anything)
	logger.On("Printf", mock.Anything, mock.Anything)
	logger.On("Warnf", mock.Anything, mock.Anything)

	installer := NewCocoapodsInstaller(cmdFactory, logger)

	// When
	err := installer.InstallPods(podArg, podCmd, "", false)

	// Then
	require.NoError(t, err)
	cmdFactory.AssertExpectations(t)
	firstInstallCmd.AssertExpectations(t)
	repoUpdateCmd.AssertExpectations(t)
	secondInstallCmd.AssertExpectations(t)
}

func Test_GivenCocoapodsErrorFinder_WhenGatewayTimeOut_ThenFindsErrors(t *testing.T) {
	expectedErrors := []string{
		"[!] Error installing boost",
		"[!] /usr/bin/curl -f -L -o /var/folders/v9/hjkgcpmn6bq99p7gvyhpq6800000gn/T/d20221018-7204-3bfs7/file.tbz https://boostorg.jfrog.io/artifactory/main/release/1.76.0/source/boost_1_76_0.tar.bz2 --create-dirs --netrc-optional --retry 2 -A 'CocoaPods/1.11.3 cocoapods-downloader/1.6.3'",
		"Transient problem",
		"curl: (22) The requested URL returned error: 504 Gateway Time-out",
	}
	errorFinder := cocoapodsCmdErrorFinder{}
	errors := errorFinder.findErrors(podInstallGatewayTimeOutError)
	require.Equal(t, expectedErrors, errors)
}

func Test_GivenCocoapodsErrorFinder_WhenBadGateway_ThenFindsErrors(t *testing.T) {
	expectedErrors := []string{
		"[!] Error installing boost",
		"[!] /usr/bin/curl -f -L -o /var/folders/v9/hjkgcpmn6bq99p7gvyhpq6800000gn/T/d20221018-3650-smj60t/file.tbz https://boostorg.jfrog.io/artifactory/main/release/1.76.0/source/boost_1_76_0.tar.bz2 --create-dirs --netrc-optional --retry 2 -A 'CocoaPods/1.11.3 cocoapods-downloader/1.6.3'",
		"Transient problem",
		"curl: (22) The requested URL returned error: 502 Bad Gateway",
	}
	errorFinder := cocoapodsCmdErrorFinder{}
	errors := errorFinder.findErrors(podInstallBadGatewayError)
	require.Equal(t, expectedErrors, errors)
}

const podInstallGatewayTimeOutError = `Installing YogaKit (1.18.1)
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
  0     0    0     0    0     0      0      0 --:--:--  0:01:15 --:--:--     0
Warning: Transient problem: HTTP error Will retry in 2 seconds. 1 retries 
Warning: left.

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:15 --:--:--     0
curl: (22) The requested URL returned error: 504 Gateway Time-out
`

const podInstallBadGatewayError = `Installing YogaKit (1.18.1)
Installing boost (1.76.0)

[!] Error installing boost
[!] /usr/bin/curl -f -L -o /var/folders/v9/hjkgcpmn6bq99p7gvyhpq6800000gn/T/d20221018-3650-smj60t/file.tbz https://boostorg.jfrog.io/artifactory/main/release/1.76.0/source/boost_1_76_0.tar.bz2 --create-dirs --netrc-optional --retry 2 -A 'CocoaPods/1.11.3 cocoapods-downloader/1.6.3'

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:05 --:--:--     0
Warning: Transient problem: HTTP error Will retry in 1 seconds. 2 retries 
Warning: left.

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:01:15 --:--:--     0
Warning: Transient problem: HTTP error Will retry in 2 seconds. 1 retries 
Warning: left.

  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:--  0:00:01 --:--:--     0
curl: (22) The requested URL returned error: 502 Bad Gateway`
