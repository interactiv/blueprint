// Copyright 2015 aikah
// License MIT

package main

import (
	"flag"
	"github.com/interactiv/blueprints/trace"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

type templateHandler struct {
	filename string
	tpl      *template.Template
	once     sync.Once
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.tpl = template.Must(template.ParseFiles(path.Join("templates", t.filename)))
	})
	t.tpl.Execute(w, r)
}

func main() {
	const (
		port string = ":8080"
	)
	// deal with cli arguments
	var addr = flag.String("addr", ":8080", "the addr of the application")
	flag.Parse()
	r := newRoom()
	r.tracer = trace.New(os.Stdout)
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	//http.Handle("/public", http.FileServer(http.Dir("./public")))
	go r.run()
	log.Println("starting server on ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
