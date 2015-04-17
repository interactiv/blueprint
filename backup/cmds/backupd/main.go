// main.go
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/interactiv/blueprints/backup"
	"github.com/matryer/filedb"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

func main() {
	var (
		interval = flag.Int("interval", 10, "interval between checks(second)")
		archive  = flag.String("archive", "archive", "path to archive location")
		dbpath   = flag.String("db", "/backupdata", "path to filedb database")
	)
	flag.Parse()
	monitor := &backup.Monitor{
		Destination: *archive,
		Archiver:    backup.ZIP,
		Paths:       make(map[string]string),
	}
	db, err := filedb.Dial(*dbpath)
	exitOnError(err)
	defer db.Close()
	col, err := db.C("paths")
	exitOnError(err)
	col.ForEach(func(_ int, data []byte) bool {
		var path path
		err := json.Unmarshal(data, &path)
		exitOnError(err)
		monitor.Paths[path.Path] = path.Hash
		return false
	})
	if len(monitor.Paths) < 1 {
		exitOnError(errors.New("no paths - use backup tool to add at least one"))
	}
	check(monitor, col)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-time.After(time.Duration(*interval) * time.Second):
			check(monitor, col)
		case <-signalChannel:
			// stop
			fmt.Println()
			log.Printf("Stopping....")
			goto stop
		}

	}
stop:
}

type path struct {
	Path string
	Hash string
}

func exitOnError(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}
func check(monitor *backup.Monitor, col *filedb.C) {
	log.Println("Checking")
	counter, err := monitor.Now()
	exitOnError(err)
	if counter > 0 {
		log.Printf("Archived %d directories\n", counter)
		// update hashes

		col.SelectEach(func(_ int, data []byte) (bool, []byte, bool) {
			var path path
			if err := json.Unmarshal(data, &path); err != nil {
				log.Println("failed to unmarshal data (skipping):", err)
				return true, data, false
			}
			path.Hash, _ = monitor.Paths[path.Path]
			newdata, err := json.Marshal(path)
			if err != nil {
				log.Println("failed to marshal data (skipping):", err)
				return true, data, false
			}
			return true, newdata, false
		})
	} else {
		log.Println("No changes.")
	}
}
