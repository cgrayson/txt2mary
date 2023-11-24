package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	TestMode = true

	config := LoadConfig()
	if config.UsersFilename != "./fixtures/users_test.json" {
		t.Errorf("Expected UsersFilename to be './fixtures/users_test.json', not %q", config.UsersFilename)
	}
	if config.MicroBlog.Token != "foo-bar-42" {
		t.Errorf("Expected Token to be 'foo-bar-42', not %q", config.MicroBlog.Token)
	}
	if config.MicroBlog.Destination != "https://foo.micro.blog/" {
		t.Errorf("Expected Destination to be 'https://foo.micro.blog/', not %q", config.MicroBlog.Destination)
	}
	if config.MicroBlog.TestDestination != "https://foo-test.micro.blog/" {
		t.Errorf("Expected TestDestination to be 'https://foo-test.micro.blog/', not %q", config.MicroBlog.TestDestination)
	}
	if config.Twitter == (TwitterConfig{}) {
		t.Errorf("Expected TwitterConfig to be populated")
	}
	if config.Twitter.TestAccount != false {
		t.Errorf("Expected TwitterConfig to default TestAccount to false")
	}
}

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
