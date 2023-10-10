package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// TestMode defaults to false but is set in the tests to load test config
var TestMode = false

type Config struct {
	UsersFilename  string
	MicroBlogToken string
}

type TwilioPayload struct {
	From       string
	Body       string // the message itself, may begin with "TEST:"
	NumMedia   int
	MediaUrl1  string
	MediaUrl2  string
	MediaUrl3  string
	MediaUrl4  string
	MediaUrl5  string
	MediaUrl6  string
	MediaUrl7  string
	MediaUrl8  string
	MediaUrl9  string
	MediaUrl10 string
}

func LoadConfig() Config {
	var config Config
	filename := "config.json"
	if TestMode {
		filename = "./fixtures/config_test.json"
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(contents, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func loadPhoneFile() map[string]string {
	config := LoadConfig()
	contents, err := os.ReadFile(config.UsersFilename)
	if err != nil {
		log.Fatal(err)
	}

	var phoneMap map[string]string
	err = json.Unmarshal(contents, &phoneMap)
	if err != nil {
		log.Fatal(err)
	}

	return phoneMap
}

func LookupPhone(phone string) string {
	phoneMap := loadPhoneFile()
	return phoneMap[phone]
}

func Format(msg string, fromPhone string) string {
	if msg == "" {
		msg = "&nbsp;"
	}
	from := LookupPhone(fromPhone)
	if from == "" {
		from = "(unknown sender)" // shouldn't happen, but just in case
	}

	return fmt.Sprintf("> %s\n\n&ndash; %s", msg, from)
}

func Twiml(msg string) string {
	return fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Response>\n    <Message>%s</Message>\n</Response>", msg)
}
