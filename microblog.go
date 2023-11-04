package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// uploadFile takes the name of the file to upload, the destination blog, and
// the Micro.blog API token, and uploads the file
func uploadFile(filename string, mpDestination string, mbToken string) (string, error) {
	mbUrl := "https://micro.blog/micropub/media?mp-destination=" + url.QueryEscape(mpDestination)

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

	request, err := http.NewRequest(http.MethodPost, mbUrl, bytes.NewReader(body.Bytes()))
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", "Bearer "+mbToken)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	// Micro.blog upload returns URL in the Location header
	return resp.Header.Get("Location"), nil
}

func postMessage(message *Message, mpDestination string, mbToken string) (string, error) {
	mbUrl := "https://micro.blog/micropub/media?mp-destination=" + url.QueryEscape(mpDestination)

	data := url.Values{}
	data.Set("h", "entry")
	data.Set("content", url.QueryEscape(message.Text))
	data.Set("category", "txt")
	for i := 0; i <= message.NumImages; i++ {
		data.Add("photo[]", message.MBImageURLs[i])
	}

	request, err := http.NewRequest(http.MethodPost, mbUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", "Bearer "+mbToken)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type microblogResponse struct {
		url     string
		preview string
	}
	mbResponse := microblogResponse{}
	err = json.Unmarshal([]byte(body), &mbResponse)
	if err != nil {
		return "", err
	}

	return mbResponse.url, nil
}

// UploadImagesToMicroBlog uploads all the images in the given Message to Micro.blog,
// updating the Message with MBImageURLs.
func UploadImagesToMicroBlog(message *Message) error {
	config := LoadConfig()

	destinationBlog := config.MicroBlog.Destination
	if strings.HasPrefix(message.Text, "TEST: ") {
		destinationBlog = config.MicroBlog.TestDestination
	}

	// this will skip if no images
	for i := 0; i < message.NumImages; i++ {
		mbUrl, err := uploadFile(message.ImageFilenames[i], destinationBlog, config.MicroBlog.Token)
		if err != nil {
			log.Printf("Error uploading file %q to blog %q", message.ImageFilenames[i], destinationBlog)
			return err
		}
		message.MBImageURLs = append(message.MBImageURLs, mbUrl)
	}

	var err error
	message.MBPostURL, err = postMessage(message, destinationBlog, config.MicroBlog.Token)
	if err != nil {
		log.Printf("Error posting message to blog %q", destinationBlog)
		return err
	}
	return nil
}
