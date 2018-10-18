package models

// BitriseConfigMap ...
type BitriseConfigMap map[string]string

// Warnings ...
type Warnings []string

// Errors ...
type Errors []string

// ScanResultModel ...
type ScanResultModel struct {
	PlatformOptionMap    map[string]OptionModel      `json:"options,omitempty" yaml:"options,omitempty"`
	PlatformConfigMapMap map[string]BitriseConfigMap `json:"configs,omitempty" yaml:"configs,omitempty"`
	PlatformWarningsMap  map[string]Warnings         `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	PlatformErrorsMap    map[string]Errors           `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// AddError ...
func (result *ScanResultModel) AddError(platform string, errorMessage string) {
	if result.PlatformErrorsMap == nil {
		result.PlatformErrorsMap = map[string]Errors{}
	}
	if result.PlatformErrorsMap[platform] == nil {
		result.PlatformErrorsMap[platform] = []string{}
	}
	result.PlatformErrorsMap[platform] = append(result.PlatformErrorsMap[platform], errorMessage)
}
