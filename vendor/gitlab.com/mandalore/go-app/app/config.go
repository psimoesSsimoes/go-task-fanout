package app

import (
	"log"
	"os"
	"path/filepath"

	viper "github.com/spf13/viper"
)

var appConfigurations = make(map[string]*ConfigurationMeta)

// ConfigurationMeta data structure contains configuration meta information and methods.
type ConfigurationMeta struct {
	v *viper.Viper
}

// LoadEnvConfig loads the configuration file for the applicatio name provided. ConfigurationMeta is always set even on error.
// The configuration must be a JSON file and it is searched in the following places
//
// For remote deployment, container images, etc, use the absolute path
//   /etc/<namespace>/<appName>.json
// For local/per-developer configurations ignore from version control and use
//   ./.{configs,config,etc}/<GOAPP_ENV>/<appName>.json
// For versioning configurations with the project for running integration tests or against test environments.
//   ./{configs,config,etc}/<GOAPP_ENV>/<appName>.json
func LoadEnvConfig(namespace, appName string, configuration interface{}) (*ConfigurationMeta, error) {
	var meta *ConfigurationMeta

	meta, found := appConfigurations[appName]
	if !found {
		meta = &ConfigurationMeta{}
		meta.v = viper.New()

		meta.v.SetConfigName(appName)
		meta.v.SetConfigType("json")

		// specific configuration folders for server (remote machine), custom (local machine) and dev (included on source code)
		if namespace != "" {
			meta.v.AddConfigPath("/etc/" + namespace)
		}

		env := GetEnvironment()
		meta.v.AddConfigPath(".configs/" + env)
		meta.v.AddConfigPath(".config/" + env)
		meta.v.AddConfigPath(".etc/" + env)
		meta.v.AddConfigPath("configs/" + env)
		meta.v.AddConfigPath("config/" + env)
		meta.v.AddConfigPath("etc/" + env)

		if err := meta.v.ReadInConfig(); err != nil {
			return meta, err
		}

		appConfigurations[appName] = meta
	}

	if err := meta.v.Unmarshal(configuration); err != nil {
		return meta, err
	}

	return meta, nil
}

// LoadConfig loads the configuration file for the applicatio name provided. ConfigurationMeta is always set even on error.
// The configuration must be a JSON file and it is searched in the following places
//
// For remote deployment, container images, etc, use the absolute path
//   /etc/<namespace>/<GOAPP_ENV>.<appName>.json
// For local/per-developer configurations ignore from version control and use
//   ./.{configs,config,etc}/<GOAPP_ENV>.<appName>.json
// For versioning configurations with the project for running integration tests or against test environments.
//   ./{configs,config,etc}/<GOAPP_ENV>.<appName>.json
func LoadConfig(namespace, appName string, configuration interface{}) (*ConfigurationMeta, error) {
	var meta *ConfigurationMeta

	meta, found := appConfigurations[appName]
	if !found {
		meta = &ConfigurationMeta{}
		meta.v = viper.New()

		meta.v.SetConfigName(GetEnvironment() + "." + appName)
		meta.v.SetConfigType("json")

		// specific configuration folders for server (remote machine), custom (local machine) and dev (included on source code)
		if namespace != "" {
			meta.v.AddConfigPath("/etc/" + namespace)
		}
		meta.v.AddConfigPath(".configs")
		meta.v.AddConfigPath(".config")
		meta.v.AddConfigPath(".etc")
		meta.v.AddConfigPath("configs")
		meta.v.AddConfigPath("config")
		meta.v.AddConfigPath("etc")

		if err := meta.v.ReadInConfig(); err != nil {
			return meta, err
		}

		appConfigurations[appName] = meta
	}

	if err := meta.v.Unmarshal(configuration); err != nil {
		return meta, err
	}

	return meta, nil
}

// LoadConfigFromPath loads a config except you can override the default paths the config is searched in.
func LoadConfigFromPath(appName string, path string, configuration interface{}) (*ConfigurationMeta, error) {
	var meta *ConfigurationMeta

	meta, found := appConfigurations[appName]
	if !found {
		meta = &ConfigurationMeta{}
		meta.v = viper.New()

		meta.v.SetConfigName(GetEnvironment() + "." + appName)
		meta.v.SetConfigType("json")

		meta.v.AddConfigPath(path)

		if err := meta.v.ReadInConfig(); err != nil {
			return meta, err
		}

		appConfigurations[appName] = meta
	}

	if err := meta.v.Unmarshal(configuration); err != nil {
		return meta, err
	}

	return meta, nil
}

// LoadConfigFile loads a config except you can override the default paths the config is searched in.
func LoadConfigFile(path string, configuration interface{}) (*ConfigurationMeta, error) {
	meta := &ConfigurationMeta{}

	meta.v = viper.New()
	meta.v.SetConfigFile(path)
	if err := meta.v.ReadInConfig(); err != nil {
		return meta, err
	}

	if err := meta.v.Unmarshal(configuration); err != nil {
		return meta, err
	}

	return meta, nil
}

// GetConfigFile returns the path to the configuration file used.
func (config *ConfigurationMeta) GetConfigFile() string {
	return config.v.ConfigFileUsed()
}

// GetConfigFolder returns the dirname of the config file path.
func (config *ConfigurationMeta) GetConfigFolder() string {
	fname := config.v.ConfigFileUsed()

	return filepath.Dir(fname)
}

// GetEnvironment returns GOAPP_ENV environment variable or 'local' if not defined.
func GetEnvironment() string {
	env := os.Getenv("GOAPP_ENV")

	if env == "" {
		env = "dev"
	}

	return env
}

// GetWorkingDir returns the current working directory
func GetWorkingDir() string {
	dir, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	return dir
}
