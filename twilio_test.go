package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

const testUrl = "/2010-04-01/Accounts/AC123/Messages/MM0123/Media/ME2fe37blahblah"

const testTwilioMsg = "{ \"ToCountry\": \"US\", \"MediaContentType0\": \"image/jpeg\", \"ToState\": \"AL\", \"SmsMessageSid\": \"MM0123\", \"NumMedia\": \"1\", \"ToCity\": \"\", \"FromZip\": \"78765\", \"SmsSid\": \"MM0123\", \"FromState\": \"TX\", \"SmsStatus\": \"received\", \"FromCity\": \"AUSTIN\", \"Body\": \"Here is another pic\", \"FromCountry\": \"US\", \"To\": \"+12055551212\", \"ToZip\": \"\", \"NumSegments\": \"1\", \"MessageSid\": \"MM0123\", \"AccountSid\": \"AC123\", \"From\": \"+15125551212\", \"MediaUrl0\": \"https://api.twilio.com/2010-04-01/Accounts/AC123/Messages/MM0123/Media/ME456\", \"ApiVersion\": \"2010-04-01\" }"

func cleanupDownload(filename string) {
	if err := os.Remove(filename); err != nil {
		log.Fatal("error removing file: " + filename)
	}
}

func TestGetTwilio(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("\n")) // just return a newline to be saved in the temp file
	}))
	defer server.Close()

	filename, _ := GetTwilioImage(server.URL + testUrl)
	defer cleanupDownload(filename)

	if filename != "ME2fe37blahblah_temp.jpg" {
		t.Errorf("Expected 'ME2fe37blahblah_temp.jpg', got %s", filename)
	}
}

// TestUnmarshalling was written more to make sure I understood unmarshalling than actual coverage
func TestUnmarshalling(t *testing.T) {
	testString := "{\"NumMedia\": \"2\", \"MediaUrl1\": \"https://url1.com\", \"MediaUrl9\": \"https://url9.com\"}"
	var twilioPayload TwilioPayload

	err := json.Unmarshal([]byte(testString), &twilioPayload)
	if err != nil {
		log.Fatal(err)
	}

	if twilioPayload.NumMedia != "2" {
		t.Errorf("expected 2 (got %q)", twilioPayload.NumMedia)
	}
	if twilioPayload.MediaUrl1 != "https://url1.com" {
		t.Errorf("expected url1.com, not %q", twilioPayload.MediaUrl1)
	}
	if twilioPayload.MediaUrl9 != "https://url9.com" {
		t.Errorf("expected url9.com, not %q", twilioPayload.MediaUrl1)
	}
}

func TestParseTwilioWebhook(t *testing.T) {
	TestMode = true
	data := url.Values{}
	data.Set("From", "+15125551212")
	data.Set("Body", "Here is another pic")
	data.Set("NumMedia", "1")
	data.Add("MediaUrl0", "https://api.twilio.com/2010-04-01/Accounts/AC123/Messages/MM0123/Media/ME456")
	message := ParseTwilioWebhook(data)

	if message.From != "Gon" {
		t.Errorf("expected Message.From of 'Gon', got %q", message.From)
	}
	if message.Text != "Here is another pic" {
		t.Errorf("expected Message.Text of 'Here is another pic', got %q", message.Text)
	}

	if message.NumImages != 1 {
		t.Errorf("expected Message.NumImages of 1, got %d", message.NumImages)
	}
	if len(message.TwilioImageURLs) != 1 {
		t.Errorf("expected 1 Message.TwilioImageURLs, got %d", len(message.TwilioImageURLs))
	}
	if message.TwilioImageURLs[0] != "https://api.twilio.com/2010-04-01/Accounts/AC123/Messages/MM0123/Media/ME456" {
		t.Errorf("expected Message.TwilioImageURL to be set, got %q", message.TwilioImageURLs[0])
	}
}

func TestDownloadTwilioImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("\n")) // just return a newline to be saved in the temp file
	}))
	defer server.Close()

	message := Message{
		NumImages:       1,
		TwilioImageURLs: []string{server.URL + "/2010-04-01/Accounts/AC123/Messages/MM0123/Media/ME456"},
	}

	err := DownloadTwilioImages(&message)
	if err != nil {
		t.Errorf("expected no error, got %q", err)
	}

	if len(message.ImageFilenames) != 1 {
		t.Errorf("expected 1 ImageFilename, got %d", len(message.ImageFilenames))
	}
	if message.ImageFilenames[0] != "ME456_temp.jpg" {
		t.Errorf("expected ImageFilename to be 'ME456_temp.jpg', got %q", message.ImageFilenames[0])
	}

	cleanupDownload(message.ImageFilenames[0])
}
