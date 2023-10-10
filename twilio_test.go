package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testUrl = "/2010-04-01/Accounts/id1/Messages/id2/Media/ME2fe37blahblah"

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
	if filename != "ME2fe37blahblah_temp.jpg" {
		t.Errorf("Expected 'ME2fe37blahblah_temp.jpg', got %s", filename)
	} else {
		if err := os.Remove(filename); err != nil {
			t.Fatal("error removing file: " + filename)
		}
	}
}
