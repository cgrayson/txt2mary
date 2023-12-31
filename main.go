package main

import (
	"fmt"
	"github.com/honeybadger-io/honeybadger-go"
	"io"
	"log"
	"net/http"
	"os"
)

type Message struct {
	From            string
	Text            string
	NumImages       int
	TwilioImageURLs []string
	ImageFilenames  []string
	MBImageURLs     []string
	MBPostURL       string
	TwitterMediaIds []string
	TwitterPostURL  string
}

var config Config
var Version = "development"

func post(message *Message) error {
	// download images, if there are any
	if message.NumImages > 0 {
		err := DownloadTwilioImages(message)
		if err != nil {
			log.Printf("error downloading from Twilio")
			return err
		}
	}

	// post the message to Micro.blog, if it's configured
	if config.MicroBlog != (MicroBlogConfig{}) {
		err := UploadMessageToMicroBlog(message)
		if err != nil {
			log.Printf("error posting message to Micro.blog")
			return err
		}
	} else {
		log.Printf("no configuration for Micro.blog - skipping")
	}

	// post the message to Twitter, if it's configured
	if config.Twitter != (TwitterConfig{}) {
		err := UploadMessageToTwitter(message)
		if err != nil {
			log.Printf("error posting message to Twitter")
			return err
		}
	} else {
		log.Printf("no configuration for Twitter - skipping")
	}
	return nil
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, "ok")
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("error reading parsing form data: %s\n", err)
	}

	message := ParseTwilioWebhook(r.PostForm)

	// check for an unrecognized sender
	if message.From == "" {
		log.Printf("message from unrecognized number; returning")
		_, err = io.WriteString(w, Twiml("your number is not allowed to text here"))
		if err != nil {
			log.Printf("error writing twiml response")
		}
		return
	}

	err = post(&message)
	if err != nil && config.HoneybadgerAPIKey != "" {
		log.Printf("notifying Honeybadger of err: %s\n", err)
		_, _ = honeybadger.Notify(err)
	}

	// always respond to Twilio (with rose-tinted message)
	_, err = io.WriteString(w, Twiml(fmt.Sprintf("message posted %s", message.MBPostURL)))
	if err != nil {
		log.Printf("error writing twiml response")
	}

	RemoveTwilioImages(message)

	log.Printf("done processing message from %s, with %d images: %q\n", message.From, message.NumImages, message.Text)
}

func main() {
	config = LoadConfig()

	if config.HoneybadgerAPIKey != "" {
		honeybadger.Configure(honeybadger.Configuration{APIKey: config.HoneybadgerAPIKey})
		defer honeybadger.Monitor() // reports unhandled panics
	}

	if config.Logfile != "stderr" && config.Logfile != "" {
		file, err := os.OpenFile(config.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("error creating logfile %q: %s", config.Logfile, err)
		}
		log.SetOutput(file)
	}
	log.Printf("config loaded; version %q listening on %s%s", Version, config.Server, config.ServerRoute)

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc(config.ServerRoute, handler)
	log.Fatal(http.ListenAndServe(config.Server, nil))
}
