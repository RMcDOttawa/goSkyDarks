/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the status of the session described in the state file",
	Long:  `Displays the saved state, giving the number frames requested, those already captured, and the remaining work.`,
	Run: func(cmd *cobra.Command, args []string) {
		//if Debug || Verbosity >= 1 {
		//	fmt.Println("Status command entered")
		//}
	},
}

func init() {
	fmt.Println("Status Init")
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
