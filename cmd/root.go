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

		tasks, _ := lib.ParseTasks("test.orca")
		g := lib.NewGraph(tasks)
		g.AddNodes()
		lib.ParseDependencies("test.orca", g)
		g.CreateTopologicalLayers()

		jsonData, _ := json.MarshalIndent(g, "", "  ")
		fmt.Println(string(jsonData))
		lib.ExecuteDAG(g)

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error using orca: '%s'", err)
		os.Exit(1)
	}
}
