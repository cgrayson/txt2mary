package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

func UploadFileToMicroBlog(filename string, isTest bool, mbToken string) {
	baseUrl := "https://micro.blog/micropub/media"

	mpDestination := "https://maryg.micro.blog/"

	// send test posts to the test blog
	if isTest {
		mpDestination = "https://maryg-test.micro.blog/"
	}

	mbUrl := baseUrl + "?mp-destination=" + url.QueryEscape(mpDestination)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename) // *must* be "file"

	_, err = io.Copy(fw, file)
	if err != nil {
		log.Fatal(err)
	}
	_ = writer.Close()

	request, err := http.NewRequest("POST", mbUrl, bytes.NewReader(body.Bytes()))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Add("Authorization", "Bearer "+mbToken)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("status: %s, body: %q\n", resp.Status, responseBody)
}
