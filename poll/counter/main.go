package main

import (
	"flag"
	"fmt"
	"github.com/iron-io/iron_go/mq"
	"github.com/joeshaw/envdecode"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const updateDuration = 1 * time.Second

var (
	fatarErr        error
	mongoConnection struct {
		String string `env:"MONGODB,required"`
		DbName string `env:"SP_DBNAME,required"`
	}
	counts     map[string]int
	countsLock sync.Mutex
)

func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatarErr = e
}

func main() {
	defer func() {
		if fatarErr != nil {
			os.Exit(1)
		}
	}()
	log.Println("Connecting to the database...")
	envdecode.Decode(&mongoConnection)
	db, err := mgo.Dial(mongoConnection.String)
	if err != nil {
		fatal(err)
		return
	}
	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()
	polldata := db.DB(mongoConnection.DbName).C("polls")
	log.Println("Connecting to queue")
	q := mq.New("votes")
	if err != nil {
		fatal(err)
		return
	}
	go func() {
		for {
			message, err := q.Get()
			if err != nil {
				fmt.Println("error getting message: ", err)
			} else {
				fmt.Printf("new vote: \n%+v\n", message)
				countsLock.Lock()
				if counts == nil {
					counts = make(map[string]int)
				}
				vote := message.Body
				counts[vote]++
				message.Delete()
				countsLock.Unlock()
			}
		}

	}()
	log.Print("Waiting for votes on queue...")
	var updater *time.Timer
	updater = time.AfterFunc(updateDuration, func() {
		countsLock.Lock()
		defer countsLock.Unlock()
		if len(counts) == 0 {
			log.Println("No new votes,skipping database update")
		} else {
			log.Println("Updating database...")
			log.Println(counts)
			ok := true
			for option, count := range counts {
				sel := bson.M{"options": bson.M{"$in": []string{option}}}
				up := bson.M{"$inc": bson.M{"results." + option: count}}
				if _, err := polldata.UpdateAll(sel, up); err != nil {
					log.Println("failed to update:", err)
					ok = false
				}
			}
			if ok {
				log.Println("Finished updating database....")
				counts = nil // reset counts
			}
		}
		updater.Reset(updateDuration)
	})
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-termChan:
			updater.Stop()
			return
		}
	}
}
