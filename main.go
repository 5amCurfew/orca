package main

import (
	"os"

	lib "github.com/5amCurfew/orca/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "0.5.0"

func main() {
	Execute()
}

var rootCmd = &cobra.Command{
	Use:     "orca [PATH_TO_DAG_FILE]",
	Version: version,
	Short:   "orca - bash orchestrator",
	Long:    `orca is a bash command orchestrator that can be used to run terminal commands in a directed acyclic graph`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := "dag.yml"
		if len(args) > 0 {
			cfgPath = args[0]
		}

		g, err := lib.NewGraph(cfgPath)
		if err != nil {
			log.Fatalf("Error initialising graph %s: %s", cfgPath, err)
		}
		g.Execute()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("[INIT] error using orca: %s", err)
		os.Exit(1)
	}
}
