package main

import (
	"fmt"
	"io"
	"io/ioutil"
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
}

var config Config

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %q\n", err)
	}

	message := ParseTwilioWebhook(string(body))

	// check for an unrecognized sender
	if message.From == "" {
		_, err = io.WriteString(w, Twiml("your number is not allowed to text here"))
		if err != nil {
			log.Fatal("Error writing twiml response")
		}
		return
	}

	// download & upload images, if there are any
	if message.NumImages > 0 {
		err := DownloadTwilioImages(&message)
		if err != nil {
			log.Fatal("Error downloading from Twilio; exiting")
		}

		err = UploadImagesToMicroBlog(&message)
		if err != nil {
			log.Fatal("Error uploading images to Micro.blog; exiting")
		}
	}

	// post the message to Micro.blog
	err = UploadMessageToMicroBlog(&message)
	if err != nil {
		log.Fatal("Error posting message to Micro.blog; exiting")
	}

	// respond to Twilio
	_, err = io.WriteString(w, Twiml(fmt.Sprintf("message posted to %s", message.MBPostURL)))
	if err != nil {
		log.Fatal("Error writing twiml response")
	}

	log.Printf("message posted from %s, with %d images: %q\n", message.From, message.NumImages, message.Text)
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

	http.HandleFunc(config.ServerRoute, handler)
	log.Fatal(http.ListenAndServe(config.Server, nil))
}
