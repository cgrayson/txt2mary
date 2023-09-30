package main

import (
	"encoding/json"
	"log"
	"testing"
)

func TestLookup(t *testing.T) {
	TestMode = true

	var tests = []struct {
		phoneNum     string
		expectedName string
	}{
		{phoneNum: "+15125551212", expectedName: "Dorothea"},
		{phoneNum: "+15125551213", expectedName: "Celia"},
		{phoneNum: "+15125551214", expectedName: ""},
	}

	for _, test := range tests {
		actual := LookupPhone(test.phoneNum)
		if actual != test.expectedName {
			t.Errorf("LookupPhone(%s) != %q", test.phoneNum, test.expectedName)
		}
	}
}

func TestFormat(t *testing.T) {
	TestMode = true

	var tests = []struct {
		msg         string
		phoneNum    string
		expectedMsg string
	}{
		{msg: "Thinking of you!", phoneNum: "+15125551212", expectedMsg: "> Thinking of you!\n\n&ndash; Dorothea"},
		{msg: "", phoneNum: "+15125551213", expectedMsg: "> &nbsp;\n\n&ndash; Celia"},
		{msg: "Hello", phoneNum: "+15125551214", expectedMsg: "> Hello\n\n&ndash; (unknown sender)"},
	}

	for _, test := range tests {
		actual := Format(test.msg, test.phoneNum)
		if actual != test.expectedMsg {
			t.Errorf("Format(%s, %s) != %q (%q)", test.msg, test.phoneNum, test.expectedMsg, actual)
		}
	}
}

// TestUnmarshalling is a test function written more to make sure I understood unmarshalling than actual coverage
func TestUnmarshalling(t *testing.T) {
	testString := "{\"NumMedia\": 2, \"MediaUrl1\": \"https://url1.com\", \"MediaUrl10\": \"https://url10.com\"}"
	var twilioPayload TwilioPayload

	err := json.Unmarshal([]byte(testString), &twilioPayload)
	if err != nil {
		log.Fatal(err)
	}

	if twilioPayload.NumMedia != 2 {
		t.Errorf("expected 2 (got %d)", twilioPayload.NumMedia)
	}
	if twilioPayload.MediaUrl1 != "https://url1.com" {
		t.Errorf("expected url1.com, not %q", twilioPayload.MediaUrl1)
	}
	if twilioPayload.MediaUrl10 != "https://url10.com" {
		t.Errorf("expected url10.com, not %q", twilioPayload.MediaUrl1)
	}
	//fmt.Printf("and here we are, with: %q\n", twilioPayload)
}
