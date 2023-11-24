package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// TestMode defaults to false but is set in the tests to load test config
var TestMode = false

type MicroBlogConfig struct {
	Token           string
	Destination     string
	TestDestination string
}

type TwitterConfig struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

type Config struct {
	Logfile           string
	Server            string
	ServerRoute       string
	UsersFilename     string
	HoneybadgerAPIKey string
	MicroBlog         MicroBlogConfig
	Twitter           TwitterConfig
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

	// ensure the configured users file is present
	_, err = os.Stat(config.UsersFilename)
	if err != nil {
		log.Fatalf("error checking users file: %s\n", err)
	}

	return config
}

func loadUsersFile() map[string]string {
	config := LoadConfig()
	contents, err := os.ReadFile(config.UsersFilename)
	if err != nil {
		log.Fatalf("error loading users file %q: %s\n", config.UsersFilename, err)
	}

	var phoneMap map[string]string
	err = json.Unmarshal(contents, &phoneMap)
	if err != nil {
		log.Fatalf("error parsing phone file %q: %s\n", config.UsersFilename, err)
	}

	return phoneMap
}

func LookupPhone(phone string) string {
	phoneMap := loadUsersFile()
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
