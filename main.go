package main

import (
	"log"
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
var TestTwilioMsg string

func main() {
	// load this at the start of each request
	config = LoadConfig()

	/*
		- listen for webhook post from Twilio
		- parse webhook post from Twilio (-> create Message, with From, Text, & Twilio Image URLs)
		- download any images to local temp files (-> add Image Filenames to Message)
		- upload any images to Micro.blog (-> add MB Image URLs to Message)
		- post actual message to Micro.blog (-> add MB Post URL to Message)
		- upload images & post message to Twitter
		- respond to Twilio
	*/
	message := ParseTwilioWebhook(TestTwilioMsg)

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

	err := UploadMessageToMicroBlog(&message)
	if err != nil {
		log.Fatal("Error posting message to Micro.blog; exiting")
	}
}
