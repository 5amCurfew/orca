package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		g := lib.NewGraph()
		g.DependOn("cake", "eggs")
		g.DependOn("cake", "flour")
		g.DependOn("eggs", "chickens")
		g.DependOn("flour", "grain")
		g.DependOn("chickens", "grain")
		g.DependOn("grain", "soil")
		g.DependOn("grain", "water")
		g.DependOn("chickens", "water")

		for i, layer := range lib.Sort(*g) {
			fmt.Printf("%d: %s\n", i, strings.Join(layer, ", "))
		}

		jsonData, _ := json.MarshalIndent(g, "", "  ")
		fmt.Println(string(jsonData))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error using orca: '%s'", err)
		os.Exit(1)
	}
}
