package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	lib "github.com/5amCurfew/orca/lib"
	"github.com/spf13/cobra"
)

var version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:     "orca",
	Version: version,
	Short:   "orca - lightweight bash orchestrator",
	Long:    `orca is a lightweight bash orchestrator that can be used to run terminal commands in a directed graph dependency structure`,
	Args:    cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		tasks := map[string]*lib.Task{
			"task1": {
				Name:    "task1",
				Command: `sleep 1.5 && echo "Task 1"`,
				Status:  "pending",
			},
			"task2": {
				Name:    "task2",
				Command: `sleep 3 && echo "Task 2"`,
				Status:  "pending",
			},
			"task3": {
				Name:    "task3",
				Command: `sleep 8 && echo "Task 3"`,
				Status:  "pending",
			},
			"task4": {
				Name:    "task4",
				Command: `sleep 1 && echo "Task 4"`,
				Status:  "pending",
			},
			"task5": {
				Name:    "task5",
				Command: `sleep 2 && echo "Task 5"`,
				Status:  "pending",
			},
			"task6": {
				Name:    "task6",
				Command: `sleep 2 && echo "Task 6"`,
				Status:  "pending",
			},
			"task7": {
				Name:    "task7",
				Command: `sleep 1 && echo "Task 7"`,
				Status:  "pending",
			},
			"task8": {
				Name:    "task8",
				Command: `sleep 0.5 && echo "Task 8"`,
				Status:  "pending",
			},
		}

		g := lib.NewGraph(tasks)

		nodes := make(map[string]struct{})
		for task := range tasks {
			nodes[task] = struct{}{}
		}
		g.AddNodes()
		g.DependOn("task3", "task1")
		g.DependOn("task3", "task2")
		g.DependOn("task4", "task1")
		g.DependOn("task6", "task4")
		g.DependOn("task6", "task5")
		g.DependOn("task4", "task3")
		g.DependOn("task7", "task2")
		g.DependOn("task8", "task1")

		g.CreateTopologicalLayers()

		jsonData, _ := json.MarshalIndent(g, "", "  ")
		fmt.Println(string(jsonData))

		// lib.ExecuteLayers(g.Layers, tasks)
		fmt.Println("--------")
		lib.ExecuteDAG(g)

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error using orca: '%s'", err)
		os.Exit(1)
	}
}
