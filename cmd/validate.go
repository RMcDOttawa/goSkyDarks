/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"goskydarks/config"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the config file",
	Long:  `Validates the config file and displays all the settings.  No capture is performed.`,
	Run:   RunValidateCommand,
}

func RunValidateCommand(_ *cobra.Command, _ []string) {
	config.ShowAllSettings()
}

func init() {
	fmt.Println("Validate Init")
	rootCmd.AddCommand(validateCmd)
}
