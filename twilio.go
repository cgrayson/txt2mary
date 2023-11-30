package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type TwilioPayload struct {
	From      string
	Body      string // the message itself, may begin with "TEST:"
	NumMedia  string
	MediaUrl0 string
	MediaUrl1 string
	MediaUrl2 string
	MediaUrl3 string
	MediaUrl4 string
	MediaUrl5 string
	MediaUrl6 string
	MediaUrl7 string
	MediaUrl8 string
	MediaUrl9 string
}

// ParseTwilioWebhook parses webhook post from Twilio,
// returning a Message populated with From, Text, & TwilioImageURLs
func ParseTwilioWebhook(formData map[string][]string) Message {
	msg := Message{
		From: LookupPhone(formData["From"][0]),
		Text: formData["Body"][0],
	}

	var err error
	numMedia := formData["NumMedia"][0]
	msg.NumImages, err = strconv.Atoi(numMedia)
	if err != nil {
		log.Printf("error converting NumMedia value from Twilio to int: %q", numMedia)
	}

	for i := 0; i < msg.NumImages; i++ {
		parameterName := fmt.Sprintf("MediaUrl%d", i)
		mediaUrl := formData[parameterName][0]
		msg.TwilioImageURLs = append(msg.TwilioImageURLs, mediaUrl)
	}

	log.Printf("received twilio post from %s, with %d images: %q", msg.From, msg.NumImages, msg.Text)
	return msg
}

// GetTwilioImage downloads the image file at the given URL,
// saves it to a filename based on the URL, and returns that filename.
func GetTwilioImage(url string) (string, error) {
	// get last ID from URL for filename
	pieces := strings.Split(url, "/")
	filename := pieces[len(pieces)-1] + "_temp.jpg"

	response, err := http.Get(url)
	if err != nil {
		return filename, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return filename, errors.New(fmt.Sprintf("Bad response code: %d", response.StatusCode))
	}

	file, err := os.Create(filename)
	if err != nil {
		return filename, err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return filename, err
	}

	return filename, nil
}

func DownloadTwilioImages(msg *Message) error {
	for i := 0; i < msg.NumImages; i++ {
		filename, err := GetTwilioImage(msg.TwilioImageURLs[i])
		if err != nil {
			log.Printf("error downloading %q", msg.TwilioImageURLs[i])
			return err
		}
		msg.ImageFilenames = append(msg.ImageFilenames, filename)
		log.Printf("downloaded image %q from Twilio\n", filename)
	}
	return nil
}

func RemoveTwilioImages(msg Message) {
	for _, filename := range msg.ImageFilenames {
		err := os.Remove(filename)
		if err != nil {
			log.Printf("error removing file %q: %s\n", filename, err)
		} else {
			log.Printf("removed file %q\n", filename)
		}
	}
}
