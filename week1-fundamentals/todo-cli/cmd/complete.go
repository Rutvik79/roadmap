/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"
	"todo-cli/tasks"

	"github.com/spf13/cobra"
)

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:   "complete [id]",
	Short: "Set the Completed Status of the task to true",
	Long: `The Completed Status of the task which has the ID = id (user provided)
	will be set to true if found, else and error will be displayed to the console`,
	Args: cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Print("Id is not valid", err)
			return
		}

		err2 := tasks.CompletedTask(id)
		if err2 != nil {
			fmt.Println("Error marking the task as Completed", err2)
			return
		}

		fmt.Println("Task marked as Completed Successfully!")
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
