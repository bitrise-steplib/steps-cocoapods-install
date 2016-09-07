package logger

import (
	"fmt"
	"os"
)

// Fail ...
func Fail(format string, v ...interface{}) {
	errorMsg := fmt.Sprintf(format, v...)
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", errorMsg)
	os.Exit(1)
}

// Warn ...
func Warn(format string, v ...interface{}) {
	errorMsg := fmt.Sprintf(format, v...)
	fmt.Printf("\x1b[33;1m%s\x1b[0m\n", errorMsg)
}

// Info ...
func Info(format string, v ...interface{}) {
	fmt.Println()
	errorMsg := fmt.Sprintf(format, v...)
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", errorMsg)
}

// Details ...
func Details(format string, v ...interface{}) {
	errorMsg := fmt.Sprintf(format, v...)
	fmt.Printf("  %s\n", errorMsg)
}

// Done ...
func Done(format string, v ...interface{}) {
	errorMsg := fmt.Sprintf(format, v...)
	fmt.Printf("  \x1b[32;1m%s\x1b[0m\n", errorMsg)
}

// Configs ...
func Configs(sourceRootPath, podfilePath, systemCocoapodsVersion string, isVerbose bool) {
	Info("Configs:")
	Details("* source_root_path: %s", sourceRootPath)
	Details("* podfile_path: %s", podfilePath)
	Details("* system_cocoapods_version: %s", systemCocoapodsVersion)
	Details("* verbose: %v", isVerbose)
}
