package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// loadConfig initializes and loads configuration automatically
func loadConfig() {
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

	// Set up viper to look for configuration
	viper.SetConfigName(".tftldr") // config file name without extension
	viper.SetConfigType("yml")     // YAML format
	viper.AddConfigPath(".")       // Look for config in the working directory

	// Try to read configuration
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if no config file is found - we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Warning: Error reading config file: %v\n", err)
		}
	}
}
