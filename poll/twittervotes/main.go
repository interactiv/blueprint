package main

import (
	"encoding/json"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/iron-io/iron_go/mq"
	"github.com/joeshaw/envdecode"
	"gopkg.in/mgo.v2"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	authClient    *oauth.Client
	creds         *oauth.Credentials
	conn          net.Conn
	reader        io.ReadCloser
	authSetupOnce sync.Once
	httpClient    *http.Client
	db            *mgo.Session
	dbName        struct {
		String string `env:"SP_DBNAME,required"`
	}
	queue = mq.New("votes")
)

func main() {
	var (
		stoplock   sync.Mutex
		stop       = false
		stopChan   = make(chan struct{}, 1)
		signalChan = make(chan os.Signal, 1)
	)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	if err := dialdb(); err != nil {
		log.Fatalln("Failed to connect mongodb:", err)
	}
	defer closedb()
	// start things
	votes := make(chan string) //chan for votes
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				break
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}

type poll struct {
	Options []string
}

func loadOptions() ([]string, error) {
	var (
		options []string
		p       poll
	)
	(&sync.Once{}).Do(func() {
		err := envdecode.Decode(&dbName)
		if err != nil {
			log.Fatal("cant find env variable SP_DBNAME", err)
		}
	})
	iter := db.DB(dbName.String).C("polls").Find(nil).Iter()
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

type tweet struct {
	Text string
}

// readFromTwitter reads tweets from twitter into a channel
func readFromTwitter(votes chan<- string) {
	var (
		options []string
		err     error
		u       *url.URL
		req     *http.Request
		resp    *http.Response
		reader  io.ReadCloser
		decoder *json.Decoder
	)
	if options, err = loadOptions(); err != nil {
		log.Println("failed to load options:", err)
		return
	}
	if u, err = url.Parse("https://stream.twitter.com/1.1/statuses/filter.json"); err != nil {
		log.Println("failed to create filter request:", err)
		return
	}
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	if req, err = http.NewRequest("POST", u.String(), strings.NewReader(query.Encode())); err != nil {
		log.Println("failed to create filter request", err)
		return
	}
	if resp, err = makeRequest(req, query); err != nil {
		log.Println("Failed to make request: ", err)
		return
	}
	reader = resp.Body
	decoder = json.NewDecoder(reader)
	for {
		var (
			tweet tweet
		)
		if err := decoder.Decode(&tweet); err != nil {
			break
		}
		for _, option := range options {
			if strings.Contains(
				strings.ToLower(tweet.Text),
				strings.ToLower(option),
			) {
				log.Println("vote:", option)
				// send only channel
				votes <- option
			}
		}
	}
}

func startTwitterStream(stopchan <-chan struct{}, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for {
			select {
			case <-stopchan:
				log.Println("stopping Twitter...")
				return
			default:
				log.Println("Querying Twitter...")
				readFromTwitter(votes)
				log.Println(" (waiting)")
				time.Sleep(10 * time.Second) //wait before reconnecting
			}
		}
	}()
	return stoppedchan
}

func publishVotes(votes <-chan string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	go func() {
		for vote := range votes {
			_, err := queue.PushString(vote) // publish vote
			if err != nil {
				log.Println("error pushing message", err)
			}
		}
		log.Println("Publisher: Stopping")
		log.Println("Publisher: Stopped")
		stopchan <- struct{}{}
	}()
	return stopchan
}
func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

func closeConn() {
	if conn != nil {
		conn.Close()

	}
	if reader != nil {
		reader.Close()
	}
}

func setupTwitterAuth() {
	var ts struct {
		ConsumerKey    string `env:"SP_TWITTER_KEY,required"`
		ConsumerSecret string `env:"SP_TWITTER_SECRET,required"`
		AccessToken    string `env:"SP_TWITTER_ACCESSTOKEN,required"`
		AccessSecret   string `env:"SP_TWITTER_ACCESSSECRET,required"`
	}
	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}
	creds = &oauth.Credentials{
		Token:  ts.AccessToken,
		Secret: ts.AccessSecret,
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}
}

// makeRequest request twitter API
func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetupOnce.Do(func() {
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: dial,
			},
		}
	})
	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))
	return httpClient.Do(req)
}

func dialdb() error {
	var (
		err             error
		mongoConnection struct {
			String string `env:"MONGODB,required"`
		}
	)
	if err := envdecode.Decode(&mongoConnection); err != nil {
		log.Fatal(err)
	}
	log.Println("dialing mongodb: ", mongoConnection.String)
	db, err = mgo.Dial(mongoConnection.String)
	return err
}

func closedb() {
	db.Close()
	log.Println("closed database connection")
}
