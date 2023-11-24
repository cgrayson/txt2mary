package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

func createTwitterClient() (client *twittergo.Client, err error) {
	clientConfig := &oauth1a.ClientConfig{
		ConsumerKey:    config.Twitter.ConsumerKey,
		ConsumerSecret: config.Twitter.ConsumerSecret,
	}
	user := oauth1a.NewAuthorizedConfig(config.Twitter.AccessToken, config.Twitter.AccessTokenSecret)
	client = twittergo.NewClient(clientConfig, user)
	return
}

func sendMediaRequest(client *twittergo.Client, reqUrl string, params map[string]string, media []byte) (mediaResp twittergo.MediaResponse, err error) {
	var (
		req         *http.Request
		resp        *twittergo.APIResponse
		body        io.ReadWriter = bytes.NewBufferString("")
		mp          *multipart.Writer
		writer      io.Writer
		contentType string
	)
	mp = multipart.NewWriter(body)
	for key, value := range params {
		mp.WriteField(key, value)
	}
	if media != nil {
		if writer, err = mp.CreateFormField("media"); err != nil {
			return
		}
		writer.Write(media)
	}
	contentType = fmt.Sprintf("multipart/form-data;boundary=%v", mp.Boundary())
	mp.Close()
	if req, err = http.NewRequest("POST", reqUrl, body); err != nil {
		return
	}
	req.Header.Set("Content-Type", contentType)
	if resp, err = client.SendRequest(req); err != nil {
		return
	}
	err = resp.Parse(&mediaResp)
	return
}

func uploadImageToTwitter(filename string) (string, error) {
	var (
		err        error
		client     *twittergo.Client
		mediaResp  twittergo.MediaResponse
		mediaId    string
		mediaBytes []byte
	)
	client, err = createTwitterClient()
	if err != nil {
		log.Printf("error creating Twitter (v1) client: %s\n", err)
		return "", err
	}
	if mediaBytes, err = ioutil.ReadFile(filename); err != nil {
		log.Printf("error reading media: %s\n", err)
		return "", err
	}
	if mediaResp, err = sendMediaRequest(
		client,
		"https://upload.twitter.com/1.1/media/upload.json",
		map[string]string{
			"media_category": "tweet_image",
		},
		mediaBytes,
	); err != nil {
		log.Printf("error sending request to Twitter (v1): %s\n", err)
		return "", err
	}
	mediaId = fmt.Sprintf("%v", mediaResp.MediaId())

	return mediaId, nil
}

func postMessageToTwitter(message *Message) (string, error) {
	const maxRetries = 5
	// this library also needs the API key & secret set in environment
	// variables $GOTWI_API_KEY & $GOTWI_API_KEY_SECRET
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           config.Twitter.AccessToken,
		OAuthTokenSecret:     config.Twitter.AccessTokenSecret,
	}

	client, err := gotwi.NewClient(in)
	if err != nil {
		log.Printf("error creating Twitter (v2) client: %s", err)
		return "", err
	}

	text := fmt.Sprintf("\"%s\"\n\n– %s", message.Text, message.From)
	input := &types.CreateInput{
		Text: gotwi.String(text),
	}

	if len(message.TwitterMediaIds) > 0 {
		input.Media = &types.CreateInputMedia{
			MediaIDs: message.TwitterMediaIds,
		}
	}

	var tweetId string
	for numTries := 0; numTries < maxRetries; numTries++ {
		log.Printf("try #%d: posting to Twitter (v2): Text: %q & MediaIDs: %v", numTries, *input.Text, message.TwitterMediaIds)
		res, err := managetweet.Create(context.Background(), client, input)
		if err != nil {
			log.Printf("error posting to Twitter (v2): Text: %q & MediaIDs: %v: %s", *input.Text, message.TwitterMediaIds, err)
		} else {
			tweetId = gotwi.StringValue(res.Data.Text)
			break
		}
	}

	if tweetId == "" {
		return "", errors.New(fmt.Sprintf("unable to post to Twitter (v2) after %d tries\n", maxRetries))
	}

	return tweetId, nil
}

func UploadMessageToTwitter(message *Message) error {
	var err error

	// only post test messages to a test account (& real messages to real account)
	if IsTestMessage(message) == config.Twitter.TestAccount {
		for _, filename := range message.ImageFilenames {
			mediaId, err := uploadImageToTwitter(filename)
			if err != nil {
				return err
			}
			log.Printf("uploaded image %q to Twitter, got mediaId %q\n", filename, mediaId)
			message.TwitterMediaIds = append(message.TwitterMediaIds, mediaId)
		}

		message.TwitterPostURL, err = postMessageToTwitter(message)
		if err != nil {
			return err
		}

		log.Printf("posted message to Twitter\n")
	} else {
		log.Printf("no Twitter account configured for this message type\n")
	}
	return nil
}
