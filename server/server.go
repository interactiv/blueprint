// Copyright 2015 mparaiso
// License MIT

package main

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"sync"
)

func main() {
	const addr string = ":8080"
	// handler par défaut
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})
	http.Handle("/template", &templateHandler{filename: "chat.html"})
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("server error", err)
	} else {
		log.Println("serving at ", addr)
	}
}

// handle templates
type templateHandler struct {
	filename string
	tpl      *template.Template
	once     sync.Once
}

// compile template once
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.tpl = template.Must(template.ParseFiles(path.Join("templates", t.filename)))
	})
	t.tpl.Execute(w, nil)
}
