/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"todo-cli/tasks"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all the tasks that have been saved by the user",
	Long: `Lists all the tasks that have been saved by the user
	using the add command with the title, this will list,
	Task ID, Task Completed Status, Task Title`,
	Run: func(cmd *cobra.Command, args []string) {
		taskList, err := tasks.ListTasks()
		if err != nil {
			fmt.Println("Error Loading tasks:", err)
			return
		}

		if len(taskList) == 0 {
			fmt.Println("No tasks found!")
			return
		}

		fmt.Println("\nYour Tasks:")
		fmt.Println("===========")
		for _, task := range taskList {
			status := "[ ]"
			if task.Completed {
				status = "[✓]"
			}
			fmt.Printf("%d. %s %s\n", task.ID, status, task.Title)
		}
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
