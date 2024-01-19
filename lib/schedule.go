package lib

import (
	"fmt"
	"log"
	"time"

	"github.com/5amCurfew/orca/util"
	"github.com/robfig/cron"
)

// Initialize the job scheduler
var dagScheduler = cron.New()

func Schedule() {
	log.Printf("Schedule initiating at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	dagFiles, _ := util.ListFiles("dags")

	for _, dagFile := range dagFiles {
		// Capture g in a closure
		func(dagFile string) {
			g, _ := NewGraph(fmt.Sprintf("dags/%s", dagFile))

			// Schedule
			if g.Schedule != "" {
				log.Printf("Schedule found for DAG %s\n", g.Name)
				dagScheduler.AddFunc(g.Schedule, func() {
					go g.Execute(time.Now())
				})
				log.Printf("Schedule initiated job for %s (%s)\n", dagFile, g.Schedule)
			}
		}(dagFile)
	}

	// Start the job scheduler
	dagScheduler.Start()
	log.Printf("Schedule(s) initiated at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	select {}
}

func UpdateSchedule() {
	log.Printf("Schedules(s) updating at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	dagScheduler.Stop()
	dagScheduler = cron.New()
	Schedule()
}
