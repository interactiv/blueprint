// main.go
package main

import (
	"flag"
	"fmt"
	"github.com/interactiv/blueprints/backup"
	"github.com/matryer/filedb"
	"log"
	"runtime/debug"
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
}

func exitOnError(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}
