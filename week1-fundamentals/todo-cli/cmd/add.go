/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"
	"todo-cli/tasks"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [task description]",
	Short: "Add a new Task",
	Long: `A Task typically consist of an "ID" (provided internally),
	 a "Title" (user provided) and a "Completed" Status 
	 which is set to false by default and the user can change it 
	 to true when the task is completed`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")
		err := tasks.AddTask(title)
		if err != nil {
			fmt.Println("Error adding task", err)
			return
		}
		fmt.Println("Task added successfully!")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
