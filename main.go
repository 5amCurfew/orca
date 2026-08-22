package main

import (
	"os"

	lib "github.com/5amCurfew/orca/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "0.6.2"

func main() {
	Execute()
}

var maxParallel int

var rootCmd = &cobra.Command{
	Use:     "orca",
	Version: version,
	Short:   "orca - bash orchestrator",
	Long:    `orca is a bash command orchestrator that can be used to run terminal commands in a directed acyclic graph`,
}

var runCmd = &cobra.Command{
	Use:   "run [PATH_TO_DAG_FILE]",
	Short: "execute the DAG",
	Long: `Execute all nodes in the DAG, respecting dependency order.

Arguments:
  PATH_TO_DAG_FILE   path to the DAG YAML file to execute (default: dag.yml)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := "dag.yml"
		if len(args) > 0 {
			file = args[0]
		}
		g, err := lib.NewGraph(file)
		if err != nil {
			log.Fatalf("Error initialising graph %s: %s", file, err)
		}
		g.Execute(maxParallel)
	},
}

var vizCmd = &cobra.Command{
	Use:   "viz [PATH_TO_DAG_FILE]",
	Short: "visualise the DAG structure in the terminal",
	Long: `Print the DAG as a dependency tree.

Arguments:
  PATH_TO_DAG_FILE   path to the DAG YAML file to visualise (default: dag.yml)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := "dag.yml"
		if len(args) > 0 {
			file = args[0]
		}
		g, err := lib.NewGraph(file)
		if err != nil {
			log.Fatalf("Error initialising graph %s: %s", file, err)
		}
		g.Visualise()
	},
}

func init() {
	runCmd.Flags().IntVarP(&maxParallel, "max-parallel", "p", 0, "maximum number of tasks to run in parallel (default: unlimited)")
	rootCmd.AddCommand(runCmd, vizCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("[INIT] error using orca: %s", err)
		os.Exit(1)
	}
}
