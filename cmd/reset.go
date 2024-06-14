/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset state file so capture starts over",
	Long:  `Remove the partially-completed information from the state file so the next "capture" command starts the entire calibration frame capture over from the beginning.`,
	Run: func(cmd *cobra.Command, args []string) {
		//if Debug || Verbosity >= 1 {
		//	fmt.Println("Reset command entered")
		//}
	},
}

func init() {
	fmt.Println("Reset Init")
	rootCmd.AddCommand(resetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
