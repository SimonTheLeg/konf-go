package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Init reads in config file and ENV variables if set.
func Init(cfgFile, konfDir string) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// Search config in home directory with name ".konfig" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".konfig")
	}
	if konfDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.Set("konfDir", home+"/.kube/konfs")
	} else {
		viper.Set("konfDir", konfDir)
	}
	// Currently there is no need to customize store and active configs individually.
	// Setting the konfDir should be fine
	viper.Set("storeDir", viper.GetString("konfDir")+"/store")
	viper.Set("activeDir", viper.GetString("konfDir")+"/active")
	viper.Set("latestKonfFile", viper.GetString("konfDir")+"/latestkonf")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	return nil
}
