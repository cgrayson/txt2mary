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

func destinationBlog(message *Message) (destination string) {
	destination = config.MicroBlog.Destination
	if strings.HasPrefix(message.Text, "TEST: ") {
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
		return "", err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename) // *must* be "file"

	_, err = io.Copy(fw, file)
	if err != nil {
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
		return "", err
	}

	// Micro.blog upload returns URL in the Location header
	return resp.Header.Get("Location"), nil
}

func postMessage(message *Message, mpDestination string) (string, error) {
	data := url.Values{}
	data.Set("h", "entry")
	data.Set("content", message.Text)
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
		return "", err
	}
	if resp.StatusCode > 202 {
		return "", errors.New(fmt.Sprintf("got a %d posting the message to Micro.blog", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var mbResponse struct {
		Url     string
		Preview string
	}
	err = json.Unmarshal(body, &mbResponse)
	if err != nil {
		return "", err
	}

	return mbResponse.Url, nil
}

// UploadImagesToMicroBlog uploads all the images in the given Message to Micro.blog,
// updating the Message with MBImageURLs.
func UploadImagesToMicroBlog(message *Message) error {
	// this will skip if no images
	for i := 0; i < message.NumImages; i++ {
		filename := message.ImageFilenames[i]
		mbUrl, err := uploadFile(filename, destinationBlog(message))
		if err != nil {
			log.Printf("Error uploading file %q to blog %q", filename, destinationBlog(message))
			return err
		}
		message.MBImageURLs = append(message.MBImageURLs, mbUrl)
		log.Printf("uploaded image %q to Micro.blog\n", filename)
	}
	return nil
}

// UploadMessageToMicroBlog sends the text, including any image URLs in the given
// Message to Micro.Blog, updating the MBPostURL with the resultant post.
func UploadMessageToMicroBlog(message *Message) error {
	var err error
	message.MBPostURL, err = postMessage(message, destinationBlog(message))
	if err != nil {
		log.Printf("Error posting message %q to blog %q: %s", message.Text, destinationBlog(message), err)
		return err
	}
	log.Printf("posted message to Micro.blog\n")
	return nil
}
