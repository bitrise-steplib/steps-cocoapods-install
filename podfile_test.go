package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsPodfileUsingSpecsRepo(t *testing.T) {
	tests := []struct {
		name           string
		podfileContent string
		expected       bool
		wantErr        bool
	}{
		{
			name:           "Empty file",
			podfileContent: "",
			expected:       false,
			wantErr:        false,
		},
		{
			name:           "Specs repo defined",
			podfileContent: repoPodfile,
			expected:       true,
			wantErr:        false,
		},
		{
			name:           "Specs repo not defined",
			podfileContent: cdnPodfile,
			expected:       false,
			wantErr:        false,
		},
		{
			name:           "Specs repo defined with quotes and whitespace",
			podfileContent: repoPodfileWithQuotes,
			expected:       true,
			wantErr:        false,
		},
		{
			name: "Other specs repo defined",
			podfileContent: otherRepoPodfile,
			expected: false,
			wantErr: false,
		},
		{
			name: "Specs repo with lowercase URL",
			podfileContent: repoPodfileLowercase,
			expected: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "Podfile")
			err := os.WriteFile(path, []byte(tt.podfileContent), 0777)
			if err != nil {
				t.Fatalf(err.Error())
			}

			actual, err := isPodfileUsingSpecsRepo(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unexpected error status. Got error: %v, want error: %v", err, tt.wantErr)
			}
			if actual != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, actual)
			}
		})
	}
}

const cdnPodfile = `
source 'https://cdn.cocoapods.org/'

platform :ios, '11.7'

def common_pods
    pod 'FirebaseAnalytics', '~> 9.4'
end
`

const repoPodfile = `
source 'https://github.com/CocoaPods/Specs.git'

platform :ios, '11.7'

def common_pods
    pod 'FirebaseAnalytics', '~> 9.4'
end
`

const repoPodfileWithQuotes = `
source "https://github.com/CocoaPods/Specs.git"  

platform :ios, '11.7'

def common_pods
    pod 'FirebaseAnalytics', '~> 9.4'
end
`

const otherRepoPodfile = `
source 'https://cdn.cocoapods.org/'
source 'https://github.com/artsy/Specs.git'

platform :ios, '11.7'

def common_pods
    pod 'FirebaseAnalytics', '~> 9.4'
end
`

const repoPodfileLowercase = `
source 'https://github.com/cocoapods/specs.git'

platform :ios, '11.7'

def common_pods
    pod 'FirebaseAnalytics', '~> 9.4'
end
`
