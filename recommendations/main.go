package main

import (
	"encoding/json"
	"github.com/interactiv/blueprints/recommendations/journey"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	runtime.GOMAXPROCS((runtime.NumCPU()))
	var (
		mainCtrl           *mainController
		googlePublicApiKey string
	)
	googlePublicApiKey = os.Getenv("GOOGLE_PUBLIC_API_KEY")
	if googlePublicApiKey == "" {
		log.Fatal("GOOGLE_PUBLIC_API_KEY environment variable not found")
	}
	mainCtrl = &mainController{ApiKey: googlePublicApiKey}
	http.HandleFunc("/journeys",
		cors(journeysHandlerFunc))
	http.HandleFunc("/recommendations",
		cors(mainCtrl.recommendationHandlerFunc))
	http.ListenAndServe(":8080", http.DefaultServeMux)

}

// cors enables C.O.R.S
func cors(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		f(w, r)
	}
}

type mainController struct {
	ApiKey string
}

// recommendationHandlerFunc handles recommendations
func (c *mainController) recommendationHandlerFunc(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	q := &journey.Query{
		Journey: strings.Split(queryValues.Get("journey"), "|"),
	}
	q.Radius, _ = strconv.Atoi(queryValues.Get("radius"))
	q.ApiKey = c.ApiKey
	q.Lat, _ = strconv.ParseFloat(queryValues.Get("lat"), 64)
	q.Lng, _ = strconv.ParseFloat(queryValues.Get("lng"), 64)
	q.CostRangeStr = queryValues.Get("cost")

	places := q.Run()
	respond(w, r, places)
}

func journeysHandlerFunc(w http.ResponseWriter, r *http.Request) {
	respond(w, r, journey.GetDefaultJourneys())
}

// respond convert the response to JSON
func respond(w http.ResponseWriter, r *http.Request, data []interface{}) error {
	publicData := []interface{}{}
	for _, d := range data {
		publicData = append(publicData, journey.Public(d))
	}
	w.Header()["Content-Type"] = []string{"application/json"}
	return json.NewEncoder(w).Encode(publicData)
}
