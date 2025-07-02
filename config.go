package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// loadConfig initializes and loads configuration from a file
func loadConfig(configPath string) {
	// Reset viper to avoid any previous configuration
	viper.Reset()

	// Set default values (will be used if no config file exists)
	viper.SetDefault("ignore.prefixes", []string{"random_", "time_"})
	viper.SetDefault("ignore.types", []string{"terraform_data", "null_resource"})

	// Default column display settings (all enabled by default)
	viper.SetDefault("columns.changeType", true)
	viper.SetDefault("columns.resourceName", true)
	viper.SetDefault("columns.changedParams", true)
	viper.SetDefault("columns.resourceType", true)
	viper.SetDefault("columns.resourceAddress", true)

	var configFile string

	// If a specific config file was provided, use that
	if configPath != "" {
		configFile = configPath
	} else {
		// Otherwise, look for a .tftldr.yml file in the current directory
		defaultConfigFile := ".tftldr.yml"
		if _, err := os.Stat(defaultConfigFile); err == nil {
			configFile = defaultConfigFile
		}
	}

	// If we found a config file, load it
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Warning: Error reading config file '%s': %v\n", configFile, err)
		} else {
			// After successful read, try to get the values
			prefixes := viper.GetStringSlice("ignore.prefixes")
			types := viper.GetStringSlice("ignore.types")

			// Ensure values are set in viper
			viper.Set("ignore.prefixes", prefixes)
			viper.Set("ignore.types", types)
		}
	}
}
