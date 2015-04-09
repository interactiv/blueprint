// auth.go
package main

import (
	"crypto/md5"
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
	"io"
	"log"
	"net/http"
	"strings"
)

// authHandler handles auth in the web app. It will check if a route needs to be secured.
// if so , checks if cookie names auth available.
type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie || cookie.Value == "" {
		// no auth
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		//other error
		panic(err.Error())
	} else {
		// success - call next
		h.next.ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
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
			m := md5.New()
			io.WriteString(m, strings.ToLower(user.Email()))
			userId := fmt.Sprintf("%x", m.Sum(nil))
			authCookieValue := objx.New(map[string]interface{}{
				"name":       user.Name(),
				"avatar_url": user.AvatarURL(),
				"email":      user.Email(),
				"userId":     userId,
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
