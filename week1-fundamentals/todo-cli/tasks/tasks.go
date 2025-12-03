package tasks

import (
	"encoding/json"
	"fmt"
	"os"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

const filename = "tasks.json"

func LoadTasks() ([]Task, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}

	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	return tasks, err
}

func SaveTasks(tasks []Task) error {
	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func AddTask(title string) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}

	newID := 1
	if len(tasks) > 0 {
		newID = tasks[len(tasks)-1].ID + 1
	}

	tasks = append(tasks, Task{
		ID:        newID,
		Title:     title,
		Completed: false,
	})

	return SaveTasks(tasks)
}

func ListTasks() ([]Task, error) {
	return LoadTasks()
}

func CompletedTask(id int) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Completed = true
			return SaveTasks(tasks)
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

func DeleteTask (id int) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return SaveTasks(tasks)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}