package models

import (
	"encoding/json"
	"fmt"
)

// OptionModel ...
type OptionModel struct {
	Title  string `json:"title,omitempty" yaml:"title,omitempty"`
	EnvKey string `json:"env_key,omitempty" yaml:"env_key,omitempty"`

	ChildOptionMap map[string]*OptionModel `json:"value_map,omitempty" yaml:"value_map,omitempty"`
	Config         string                  `json:"config,omitempty" yaml:"config,omitempty"`

	Components []string     `json:"-" yaml:"-"`
	Head       *OptionModel `json:"-" yaml:"-"`
}

// NewOption ...
func NewOption(title, envKey string) *OptionModel {
	return &OptionModel{
		Title:          title,
		EnvKey:         envKey,
		ChildOptionMap: map[string]*OptionModel{},
		Components:     []string{},
	}
}

// NewConfigOption ...
func NewConfigOption(name string) *OptionModel {
	return &OptionModel{
		ChildOptionMap: map[string]*OptionModel{},
		Config:         name,
		Components:     []string{},
	}
}

func (option *OptionModel) String() string {
	bytes, err := json.MarshalIndent(option, "", "\t")
	if err != nil {
		return fmt.Sprintf("failed to marshal, error: %s", err)
	}
	return string(bytes)
}

// IsConfigOption ...
func (option *OptionModel) IsConfigOption() bool {
	return option.Config != ""
}

// IsValueOption ...
func (option *OptionModel) IsValueOption() bool {
	return option.Title != ""
}

// IsEmpty ...
func (option *OptionModel) IsEmpty() bool {
	return !option.IsValueOption() && !option.IsConfigOption()
}

// AddOption ...
func (option *OptionModel) AddOption(forValue string, newOption *OptionModel) {
	option.ChildOptionMap[forValue] = newOption

	if newOption != nil {
		newOption.Components = append(option.Components, forValue)

		if option.Head == nil {
			// first option's head is nil
			newOption.Head = option
		} else {
			newOption.Head = option.Head
		}
	}
}

// AddConfig ...
func (option *OptionModel) AddConfig(forValue string, newConfigOption *OptionModel) {
	option.ChildOptionMap[forValue] = newConfigOption

	if newConfigOption != nil {
		newConfigOption.Components = append(option.Components, forValue)

		if option.Head == nil {
			// first option's head is nil
			newConfigOption.Head = option
		} else {
			newConfigOption.Head = option.Head
		}
	}
}

// Parent ...
func (option *OptionModel) Parent() (*OptionModel, string, bool) {
	if option.Head == nil {
		return nil, "", false
	}

	parentComponents := option.Components[:len(option.Components)-1]
	parentOption, ok := option.Head.Child(parentComponents...)
	if !ok {
		return nil, "", false
	}
	underKey := option.Components[len(option.Components)-1:][0]
	return parentOption, underKey, true
}

// Child ...
func (option *OptionModel) Child(components ...string) (*OptionModel, bool) {
	currentOption := option
	for _, component := range components {
		childOption := currentOption.ChildOptionMap[component]
		if childOption == nil {
			return nil, false
		}
		currentOption = childOption
	}
	return currentOption, true
}

// LastChilds ...
func (option *OptionModel) LastChilds() []*OptionModel {
	lastOptions := []*OptionModel{}

	var walk func(*OptionModel)
	walk = func(opt *OptionModel) {
		if len(opt.ChildOptionMap) == 0 {
			lastOptions = append(lastOptions, opt)
			return
		}

		for _, childOption := range opt.ChildOptionMap {
			if childOption == nil {
				lastOptions = append(lastOptions, opt)
				return
			}

			if childOption.IsConfigOption() {
				lastOptions = append(lastOptions, opt)
				return
			}

			if childOption.IsEmpty() {
				lastOptions = append(lastOptions, opt)
				return
			}

			walk(childOption)
		}
	}

	walk(option)

	return lastOptions
}

// RemoveConfigs ...
func (option *OptionModel) RemoveConfigs() {
	lastChilds := option.LastChilds()
	for _, child := range lastChilds {
		for _, child := range child.ChildOptionMap {
			child.Config = ""
		}
	}
}

// AttachToLastChilds ...
func (option *OptionModel) AttachToLastChilds(opt *OptionModel) {
	childs := option.LastChilds()
	for _, child := range childs {
		values := child.GetValues()
		for _, value := range values {
			child.AddOption(value, opt)
		}
	}
}

// Copy ...
func (option *OptionModel) Copy() *OptionModel {
	bytes, err := json.Marshal(*option)
	if err != nil {
		return nil
	}

	var optionCopy OptionModel
	if err := json.Unmarshal(bytes, &optionCopy); err != nil {
		return nil
	}

	return &optionCopy
}

// GetValues ...
func (option *OptionModel) GetValues() []string {
	if option.Config != "" {
		return []string{option.Config}
	}

	values := []string{}
	for value := range option.ChildOptionMap {
		values = append(values, value)
	}
	return values
}
