// Copyright 2015 mparaiso
// License MIT

package main

import (
	"log"
	"net/http"
)

func main() {
	const addr string = ":8080"
	// handler par défaut
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("server error", err)
	} else {
		log.Println("serving at ", addr)
	}
}
