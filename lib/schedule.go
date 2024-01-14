package lib

import (
	"fmt"
	"log"
	"time"

	"github.com/5amCurfew/orca/util"
	"github.com/robfig/cron"
)

func Schedule() {
	log.Printf("Schedule initiating at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// Initialize the job scheduler
	c := cron.New()

	dagFiles, _ := util.GetDagFiles()

	for _, dagFile := range dagFiles {
		// Capture g in a closure
		func(dagFile string) {
			g, _ := NewGraph(fmt.Sprintf("dags/%s", dagFile))

			// Add scheduled job for every 5 minutes
			c.AddFunc("0 */2 * * *", func() {
				go g.Execute(time.Now())
			})
		}(dagFile)
	}

	// Start the job scheduler
	c.Start()
	log.Printf("Schedule initiated at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	select {}
}
