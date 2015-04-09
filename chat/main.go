// Copyright 2015 aikah
// License MIT

package main

import (
	"flag"
	"github.com/interactiv/blueprints/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

func main() {
	const (
		port string = ":8080"
	)
	// deal with cli arguments
	var (
		addr         = flag.String("addr", "localhost:8080", "the addr of the application")
		debug        = flag.Bool("debug", false, "launch server in debug mode")
		configString = flag.String("config", os.Getenv("CHATCONFIG"), "a json string detailing the application config")
		avatar       Avatar
		r            *Room
		err          error
		config       *config
	)
	flag.Parse()
	avatar = UseFileSystemAvatar
	r = newRoom(avatar)
	config = NewConfigFromString(*configString)
	//oauth2
	gomniauth.SetSecurityKey(config.SecurityKey)
	gomniauth.WithProviders(
		github.New(
			config.Github.ClientId,
			config.Github.Secret,
			config.Github.Callback),
		google.New(
			config.Google.ClientId,
			config.Google.Secret,
			config.Google.Callback),
	)
	r.tracer = trace.New(os.Stdout)
	http.Handle("/", ServeTemplate("index.html", *debug))
	http.Handle("/chat", MustAuth(ServeTemplate("chat.html", *debug)))
	http.Handle("/login", ServeTemplate("login.html", *debug))
	http.HandleFunc("/logout", logoutHandler)
	http.Handle("/room", r)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets/"))))
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/upload", MustAuth(ServeTemplate("upload.html", *debug)))
	http.HandleFunc("/uploader", MustAuthFunc(uploadHandler))
	http.Handle("/avatars/", MustAuth(http.StripPrefix("/avatars", http.FileServer(http.Dir("./avatars/")))))
	go r.run()
	log.Println("starting server on ", *addr)
	if err = http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}

type templateHandler struct {
	filename string
	tpl      *template.Template
	once     sync.Once
	debug    bool
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !t.debug {
		t.once.Do(func() {
			t.tpl = template.Must(template.ParseFiles(path.Join("templates", t.filename)))
		})
	} else {
		t.tpl = template.Must(template.ParseFiles(path.Join("templates", t.filename)))
	}
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.tpl.Execute(w, data)
}

// ServeTemplate is a helper that returns a template handler
func ServeTemplate(templateFile string, debug bool) http.Handler {
	return &templateHandler{filename: templateFile, debug: debug}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.FormValue("userId")
	file, header, err := r.FormFile("avatarFile")
	if err != nil {
		io.WriteString(w, "error getting file from form: "+err.Error())
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		io.WriteString(w, "error reading file: "+err.Error())
		return
	}
	filename := path.Join("avatars", userId+path.Ext(header.Filename))
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	w.Header()["Location"] = []string{"/chat"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}
