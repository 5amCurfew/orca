package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	lib "github.com/5amCurfew/orca/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "0.1.7"

func main() {
	Execute()
}

var rootCmd = &cobra.Command{
	Use:     "orca [PATH_TO_DAG_FILE]",
	Version: version,
	Short:   "orca - lightweight bash orchestrator",
	Long:    `orca is a bash command orchestrator that can be used to run terminal commands in a directed acyclic graph`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})

		var cfgPath string
		if len(args) == 0 {
			// If no argument provided, look for config.json in the current directory
			log.Info("[INIT] .orca DAG file path not provided -> defaulting to dag.orca")
			cfgPath = "dag.orca"
		} else {
			cfgPath = args[0]
		}

		g, err := lib.NewGraph(cfgPath)
		if err != nil {
			log.Errorf("[INIT] %s", err)
			os.Exit(1)
		}

		jsonData, _ := json.Marshal(g)
		var gMap map[string]interface{}
		_ = json.Unmarshal(jsonData, &gMap)
		// log.WithFields(gMap).Info("Graph")

		g.Execute()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[INIT] error using orca: '%s'", err)
		os.Exit(1)
	}
}
