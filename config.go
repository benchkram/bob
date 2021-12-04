package main

import (
	"fmt"

	"github.com/Benchkram/errz"

	"github.com/fatih/structs"
	"github.com/sanity-io/litter"
	"github.com/spf13/viper"
)

// global configuration through a config file, environment vars  cli parameters.

//  Config the global config object
var GlobalConfig *config // nolint:varcheck, unused

// Create private data struct to hold config options.
// `mapstructure` => viper tags
// `struct` => fatih structs tag
type config struct {
	Verbosity  int  `mapstructure:"verbosity" structs:"verbosity"`
	CPUProfile bool `mapstructure:"cpuprofile" structs:"cpuprofile"`
	MEMProfile bool `mapstructure:"memprofile" structs:"memprofile"`
}

var defaultConfig = &config{
	Verbosity:  1,
	CPUProfile: false,
	MEMProfile: false,
}

func (c *config) AsMap() map[string]interface{} {
	return structs.Map(c)
}

func (c *config) Print() {
	litter.Dump(c)
}

// configInit must be called from the packages init() func
func configInit() {
	flags()
	bind()
	env()
}

func flags() {
	rootCmd.PersistentFlags().IntP("verbosity", "v", defaultConfig.Verbosity, "set verbosity level")
	rootCmd.PersistentFlags().Bool("cpuprofile", defaultConfig.CPUProfile, "write cpu profile to file")
	rootCmd.PersistentFlags().Bool("memprofile", defaultConfig.MEMProfile, "write memory profile to file")
}
func bind() {
	errz.Fatal(viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity")))
	errz.Fatal(viper.BindPFlag("cpuprofile", rootCmd.PersistentFlags().Lookup("cpuprofile")))
	errz.Fatal(viper.BindPFlag("memprofile", rootCmd.PersistentFlags().Lookup("memprofile")))
}
func env() {
	errz.Fatal(viper.BindEnv("verbosity", "BOB_VERBOSITY"))
	errz.Fatal(viper.BindEnv("cpuprofile", "BOB_CPU_PROFILE"))
	errz.Fatal(viper.BindEnv("memprofile", "BOB_MEM_PROFILE"))
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

func readGlobalConfig() {
	// Priority of configuration options
	// 1: CLI Parameters
	// 2: environment
	// 2: config.yaml
	// 3: defaults
	config, err := readConfig(defaultConfig.AsMap())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	GlobalConfig = config
}
