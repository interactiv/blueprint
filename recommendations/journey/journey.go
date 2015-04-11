package journey

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Journey is a trip
type Journey struct {
	Name       string
	PlaceTypes []string
}

func (journey *Journey) Public() interface{} {
	return map[string]interface{}{
		"name":    journey.Name,
		"journey": strings.Join(journey.PlaceTypes, "|"),
	}
}

// GetDefaultJourneys returns an array of journeys
func GetDefaultJourneys() []interface{} {
	return []interface{}{
		&Journey{Name: "Romantic", PlaceTypes: []string{"park", "bar",
			"movie_theater", "restaurant", "florist", "taxi_stand"}},
		&Journey{Name: "Shopping", PlaceTypes: []string{"department_store",
			"cafe", "clothing_store", "jewelry_store", "shoe_store"}},
		&Journey{Name: "Night Out", PlaceTypes: []string{"bar", "casino",
			"food", "bar", "night_club", "bar", "bar", "hospital"}},
		&Journey{Name: "Culture", PlaceTypes: []string{"museum", "cafe",
			"cemetery", "library", "art_gallery"}},
		&Journey{Name: "Pamper", PlaceTypes: []string{"hair_care",
			"beauty_salon", "cafe", "spa"}},
	}

}

type Facade interface {
	Public() interface{}
}

func Public(o interface{}) interface{} {
	if p, ok := o.(Facade); ok {
		return p.Public()
	}
	return o
}

// Place represents a place.
//
// it's worth noticing that the Place
// type embeds the googleGeometrytype,
// which allows us to represent the nested
// data as per the API, while essentially
// flattening it in our code.
type Place struct {
	*googleGeometry `json:"geometry"`
	Name            string         `json:"name"`
	Icon            string         `json:"icon"`
	Photos          []*googlePhoto `json:"photos"`
	Vicinity        string         `json:"vicinity"`
}

func (p *Place) Public() interface{} {
	return map[string]interface{}{
		"name":     p.Name,
		"icon":     p.Icon,
		"photos":   p.Photos,
		"vinicity": p.Vicinity,
		"lat":      p.Lat,
		"lng":      p.Lng,
	}
}

type googleResponse struct {
	Results []*Place `json:"results"`
}

type googleGeometry struct {
	*googleLocation `json:"location"`
}

type googleLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type googlePhoto struct {
	PhotoRef string `json:"photo_reference"`
	URL      string `json:"url"`
}

type Cost int8

const (
	_ Cost = iota
	Cost1
	Cost2
	Cost3
	Cost4
	Cost5
)

func (c Cost) String() (res string) {
	if int(c) > 5 {
		res = "invalid"
	} else {
		for i := 0; i < int(c); i++ {
			res = res + "$"
		}
	}
	return
}

func ParseCost(str string) (c Cost) {
	switch len(str) {
	case 1:
		c = Cost1
	case 2:
		c = Cost2
	case 3:
		c = Cost3
	case 4:
		c = Cost4
	case 5:
		c = Cost5
	}
	return
}

type CostRange struct {
	From Cost
	To   Cost
}

func (c *CostRange) String() string {
	return c.From.String() + "..." + c.To.String()
}

func ParseCostRange(rng string) *CostRange {
	parts := strings.Split(rng, "...")
	if len(parts) == 2 {
		return &CostRange{
			From: ParseCost(parts[0]),
			To:   ParseCost(parts[1]),
		}
	}
	return new(CostRange)
}

type Query struct {
	Lat          float64
	Lng          float64
	Journey      []string
	Radius       int
	CostRangeStr string
	ApiKey       string
}

func (q *Query) find(types string) (*googleResponse, error) {
	u := "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
	vals := make(url.Values)
	vals.Set("location", fmt.Sprintf("%g,%g", q.Lat, q.Lng))
	vals.Set("radius", fmt.Sprintf("%d", q.Radius))
	vals.Set("types", types)
	vals.Set("key", q.ApiKey)

	if len(q.CostRangeStr) > 0 {
		r := ParseCostRange(q.CostRangeStr)
		vals.Set("minprice", fmt.Sprintf("%d", int(r.From-1)))
		vals.Set("maxprice", fmt.Sprintf("%d", int(r.To-1)))
	}
	res, err := http.Get(u + "?" + vals.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response googleResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Run runs the query concurrently,and returns the results.
func (q *Query) Run() []interface{} {
	rand.Seed(time.Now().UnixNano())
	var (
		w      sync.WaitGroup
		l      sync.Mutex
		places []interface{} = make([]interface{}, len(q.Journey))
	)
	for i, r := range q.Journey {
		w.Add(1)
		go func(types string, i int) {
			defer w.Done()
			response, err := q.find(types)
			if err != nil {
				log.Println("Failed to find places:", err)
				return
			}
			if len(response.Results) == 0 {
				log.Println("No places found for", types)
				return
			}
			for _, result := range response.Results {
				for _, photo := range result.Photos {
					photo.URL = "https://maps.googleapis.com/maps/api/place/photo?" + "maxwidth=1000&photoreference=" + photo.PhotoRef + "&key=" + q.ApiKey
				}
			}
			randI := rand.Intn(len(response.Results))
			l.Lock()
			places[i] = response.Results[randI]
			l.Unlock()

		}(r, i)
	}
	w.Wait()
	return places
}
