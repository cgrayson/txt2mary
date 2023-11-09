package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
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

func getMediaUrl(payload *TwilioPayload, mediaUrlNumber int) string {
	fieldName := fmt.Sprintf("MediaUrl%d", mediaUrlNumber)
	r := reflect.ValueOf(payload)
	f := reflect.Indirect(r).FieldByName(fieldName)
	return f.String()
}

// ParseTwilioWebhook parses webhook post from Twilio,
// returning a Message populated with From, Text, & TwilioImageURLs
func ParseTwilioWebhook(payload string) Message {
	var twilioPayload TwilioPayload

	err := json.Unmarshal([]byte(payload), &twilioPayload)
	if err != nil {
		log.Printf("error unmarshalling payload %q: %v", payload, err)
		return Message{}
	}

	msg := Message{
		From: LookupPhone(twilioPayload.From),
		Text: twilioPayload.Body,
	}

	msg.NumImages, err = strconv.Atoi(twilioPayload.NumMedia)
	if err != nil {
		log.Printf("error converting NumMedia value from Twilio to int: %q", twilioPayload.NumMedia)
	}

	for i := 0; i < msg.NumImages; i++ {
		msg.TwilioImageURLs = append(msg.TwilioImageURLs, getMediaUrl(&twilioPayload, i))
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
	}
	return nil
}
