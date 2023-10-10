package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

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

//func main() {
//	espeonUrl := "https://api.twilio.com/2010-04-01/Accounts/ACdb3423fd91e1e4812a536691517ddc4d/Messages/MM54c0e36f199fb878ac1032cc53895ffb/Media/ME2473017db5e9ef5ffb8321be75db708e"
//	name, err := GetTwilioImage(espeonUrl)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("Downloaded: " + name)
//}
