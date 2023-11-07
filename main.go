package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	// load at start of each request - not performant but easy to change config
	config = LoadConfig()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %q\n", err)
	}

	message := ParseTwilioWebhook(string(body))

	// todo: handle unrecognized senders

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

	err = UploadMessageToMicroBlog(&message)
	if err != nil {
		log.Fatal("Error posting message to Micro.blog; exiting")
	}

	_, err = io.WriteString(w, Twiml(fmt.Sprintf("message posted to %s", message.MBPostURL)))
	if err != nil {
		log.Fatal("Error writing twiml response")
	}

	log.Printf("message posted from %s, with %d images: '%s'\n", message.From, message.NumImages, message.Text)
}

func main() {
	// also todo: fix port
	http.HandleFunc("/chipot-acquired", handler)
	log.Fatal(http.ListenAndServe("localhost:8088", nil))

}
