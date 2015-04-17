package main

import (
	"fmt"
	"github.com/iron-io/iron_go/mq"
	"time"
)

func main() {
	var (
		t time.Time
	)
	queue := mq.New("test_queue")
	go func(q *mq.Queue) {
		for {
			msg, _ := q.Get()
			if msg != nil {
				fmt.Printf("Message: %+v\n\n", msg)
				msg.Delete()
			}
		}
	}(queue)
	for {
		select {
		case t = <-time.After(5 * time.Second):
			queue.PushString(fmt.Sprint(t))
		}
	}
}
