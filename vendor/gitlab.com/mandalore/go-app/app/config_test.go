package app

import "testing"

// SampleConfigurationData ...
type SampleConfigurationData struct {
	Logger struct {
		Level string `mapstructure:"level"`
		Mode  string `mapstructure:"mode"`
	} `mapstructure:"logger"`
}

func TestLoadConfiguration(T *testing.T) {
	// TODO
}
