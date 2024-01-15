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

	dagFiles, _ := util.GetDagFiles()

	for _, dagFile := range dagFiles {
		// Capture g in a closure
		func(dagFile string) {
			g, _ := NewGraph(fmt.Sprintf("dags/%s", dagFile))

			// Add scheduled job for every 5 minutes
			dagScheduler.AddFunc("0 */2 * * *", func() {
				go g.Execute(time.Now())
			})
			fmt.Printf("Schedule initiated job for %s\n", dagFile)
		}(dagFile)
	}

	// Start the job scheduler
	dagScheduler.Start()

	log.Printf("Schedule(s) initiated at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	select {}
}

func UpdateSchedule() {
	log.Println("Schedules(s) updating at", time.Now().Format("2006-01-02 15:04:05"))
	dagScheduler.Stop()
	dagScheduler = cron.New()
	Schedule()
}
