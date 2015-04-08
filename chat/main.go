// Copyright 2015 aikah
// License MIT

package main

import (
	"flag"
	"fmt"
	"github.com/interactiv/blueprints/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
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
	)
	flag.Parse()
	r := newRoom()
	config := NewConfigFromString(*configString)
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
	go r.run()
	log.Println("starting server on ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
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

// fomat: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	if len(segs) != 4 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 Not Found : %s", r.URL.Path)
		return
	}
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		switch provider {
		case "google", "github":
			provider, err := gomniauth.Provider(provider)
			if err != nil {
				log.Fatalf("Error when trying to get provider %v - %v", provider, err)
			}
			loginUrl, err := provider.GetBeginAuthURL(nil, nil)
			if err != nil {
				log.Fatalf("Error when trying to GetBeginAuthURL for %v - %v ", provider, err)
			}
			w.Header().Set("Location", loginUrl)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	case "callback":
		switch provider {
		case "google", "github":
			provider, err := gomniauth.Provider(provider)
			if err != nil {
				log.Fatalf("Error when trying to complete auth for %v - %v", provider, err)
			}
			creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
			if err != nil {
				log.Fatalf("Error when trying to get user from %v - %v", provider, err)
			}
			user, err := provider.GetUser(creds)
			if err != nil {
				log.Fatalf("Error when trying to get user from %v - %v", provider, err)
			}
			authCookieValue := objx.New(map[string]interface{}{
				"name":       user.Name(),
				"avatar_url": user.AvatarURL(),
			}).MustBase64()
			http.SetCookie(w, &http.Cookie{
				Name:     "auth",
				Value:    authCookieValue,
				Path:     "/",
				HttpOnly: true,
			})
			w.Header()["Location"] = []string{"/chat"}
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Auth action %s not supported.", action)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.Header()["Location"] = []string{"/chat"}
	w.WriteHeader(http.StatusTemporaryRedirect)
}
