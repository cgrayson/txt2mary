package main

import (
	"fmt"
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

func post(message *Message) {
	// download images, if there are any
	if message.NumImages > 0 {
		err := DownloadTwilioImages(message)
		if err != nil {
			log.Printf("error downloading from Twilio")
			return
		}
	}

	// post the message to Micro.blog, if it's configured
	if config.MicroBlog != (MicroBlogConfig{}) {
		err := UploadMessageToMicroBlog(message)
		if err != nil {
			log.Printf("error posting message to Micro.blog")
			return
		}
	} else {
		log.Printf("no configuration for Micro.blog - skipping")
	}

	// post the message to Twitter, if it's configured
	if config.Twitter != (TwitterConfig{}) {
		err := UploadMessageToTwitter(message)
		if err != nil {
			log.Printf("error posting message to Twitter")
			return
		}
	} else {
		log.Printf("no configuration for Twitter - skipping")
	}
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

	post(&message)

	// always respond to Twilio (with rose-tinted message)
	_, err = io.WriteString(w, Twiml(fmt.Sprintf("message posted %s", message.MBPostURL)))
	if err != nil {
		log.Printf("error writing twiml response")
	}

	// todo: add back in file removal
	//err = os.Remove(filename)
	//if err != nil {
	//	log.Printf("Error removing file %q: %s\n", filename, err)
	//} else {
	//	log.Printf("removed file %q\n", filename)
	//}

	log.Printf("done processing message from %s, with %d images: %q\n", message.From, message.NumImages, message.Text)
}

func main() {
	config = LoadConfig()

	if config.Logfile != "stderr" && config.Logfile != "" {
		file, err := os.OpenFile(config.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("error creating logfile %q: %s", config.Logfile, err)
		}
		log.SetOutput(file)
	}
	log.Printf("config loaded, listening on %s%s", config.Server, config.ServerRoute)

	http.HandleFunc(config.ServerRoute, handler)
	log.Fatal(http.ListenAndServe(config.Server, nil))
}
