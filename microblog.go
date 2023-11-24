package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// destinationBlog takes a Message and determines which Micro.blog destination
// URL to post it too. Test messages go to the test blog, if configured.
func destinationBlog(message *Message) (destination string) {
	destination = config.MicroBlog.Destination
	if IsTestMessage(message) {
		destination = config.MicroBlog.TestDestination
	}
	return
}

func newMbRequest(mpDestination string, media bool, body io.Reader) (*http.Request, error) {
	mbUrl := "https://micro.blog/micropub"
	if media {
		mbUrl += "/media"
	}
	mbUrl += "?mp-destination=" + url.QueryEscape(mpDestination)
	request, err := http.NewRequest(http.MethodPost, mbUrl, body)
	if err != nil {
		log.Printf("error creating Micro.blog request: %s", err)
		return &http.Request{}, err
	}
	request.Header.Add("Authorization", "Bearer "+config.MicroBlog.Token)

	return request, nil
}

// uploadFile takes the name of the file to upload, the destination blog, and
// the Micro.blog API token, and uploads the file
func uploadFile(filename string, mpDestination string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("error opening file %q: %s", filename, err)
		return "", err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename) // *must* be "file"

	_, err = io.Copy(fw, file)
	if err != nil {
		log.Printf("error io-copying file %q: %s", filename, err)
		return "", err
	}
	_ = writer.Close()

	request, err := newMbRequest(mpDestination, true, bytes.NewReader(body.Bytes()))
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("error posting file %q to Micro.blog: %s", filename, err)
		return "", err
	}

	// Micro.blog upload returns URL in the Location header
	return resp.Header.Get("Location"), nil
}

func postMessage(message *Message, mpDestination string) (string, error) {
	data := url.Values{}
	data.Set("h", "entry")
	data.Set("content", fmt.Sprintf("> %s\n\n&ndash; %s", message.Text, message.From))
	data.Set("category", "txt")
	for i := 0; i < message.NumImages; i++ {
		data.Add("photo[]", message.MBImageURLs[i])
	}

	request, err := newMbRequest(mpDestination, false, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("error posting to Micro.blog: %s", err)
		return "", err
	}
	if resp.StatusCode > 202 {
		return "", errors.New(fmt.Sprintf("got status code %d posting the message to Micro.blog", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response from Micro.blog: %s", err)
		return "", err
	}

	var mbResponse struct {
		Url     string
		Preview string
	}
	err = json.Unmarshal(body, &mbResponse)
	if err != nil {
		log.Printf("error unmarshalling response from Micro.blog: %s", err)
		return "", err
	}

	return mbResponse.Url, nil
}

// UploadMessageToMicroBlog sends the text, including uploading any image in the given
// Message to Micro.Blog, updating the MBPostURL with the resultant post.
func UploadMessageToMicroBlog(message *Message) error {
	var err error

	destination := destinationBlog(message)

	// could be empty if for a test message with no TestDestination configured
	if destination != "" {
		for _, filename := range message.ImageFilenames {
			mbUrl, err := uploadFile(filename, destination)
			if err != nil {
				return err
			}
			message.MBImageURLs = append(message.MBImageURLs, mbUrl)
			log.Printf("uploaded image %q to Micro.blog\n", filename)
		}

		message.MBPostURL, err = postMessage(message, destination)
		if err != nil {
			return err
		}
		log.Printf("posted message to Micro.blog\n")
	} else {
		log.Printf("no destination blog configured for this message type\n")
	}
	return nil
}
