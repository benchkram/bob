package main

import (
	"fmt"

	"github.com/Benchkram/errz"

	"github.com/fatih/structs"
	"github.com/sanity-io/litter"
	"github.com/spf13/viper"
)

// handle global configuration through a config file, environment vars  cli parameters.

//  Config the global config object
var GlobalConfig *config // nolint:varcheck, unused

func readGlobalConfig() {
	// Priority of configuration options
	// 1: CLI Parameters
	// 2: environment
	// 2: config.yaml
	// 3: defaults
	config, err := readConfig(defaultConfig.AsMap())
	if err != nil {
		panic(err.Error())
	}
	//config.Print()

	// Set config object for main package
	GlobalConfig = config
}

var defaultConfig = &config{
	Host: "",
}

// configInit must be called from the packages init() func
func configInit() {
	// Keep cli parameters in sync with the config struct
	rootCmd.PersistentFlags().String("host", "", "hostname to listen to")

	// CLI PARMETERS
	err := viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	errz.Fatal(err)

	// ENVIRONMENT VARS
	err = viper.BindEnv("host", "HOST")
	errz.Fatal(err)
}

// Create private data struct to hold config options.
// `mapstructure` => viper tags
// `struct` => fatih structs tag
type config struct {
	Host string `mapstructure:"host" structs:"host"`
}

func (c *config) AsMap() map[string]interface{} {
	return structs.Map(c)
}

func (c *config) Print() {
	litter.Dump(c)
}

// readConfig a helper to read default from a default config object.
func readConfig(defaults map[string]interface{}) (*config, error) {
	for key, value := range defaults {
		viper.SetDefault(key, value)
	}

	//Read config from file
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	switch err.(type) {
	case viper.ConfigFileNotFoundError:
		// fmt.Printf("%s\n", aurora.Yellow("Could not find a config file"))
	default:
		return nil, fmt.Errorf("config file invalid: %w", err)
	}

	c := &config{}
	err = viper.Unmarshal(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
