# txt2mary

A Twilio-to-Micro.blog-(and-Twitter) server

This app is a special-purpose server that accepts SMS text-messages, via webhook calls from Twilio, and posts those messages to [Micro.blog](https://micro.blog/) (and Twitter*). I wrote the first version of it shortly after the death of my daughter Mary in a car crash:

> We loved sending messages to Mary, and we know a lot of her friends did, too.
> 
> I wondered if there might be some place we could still do that, but didn’t want to use a proprietary service like Twitter, Instagram, etc. Then I had the idea of a phone number that could receive SMS text messages, and how those might be published to a simple, open website. This is the result... here’s how it works. You text a message to a special phone number, and that gets posted here on this website (and on Twitter).

Like this: 

![A screenshot of the Messages app, showing a picture and text message sent, and a "message posted" reply. The picture is two cats looking out a window, with the message "just a couple bros watching the nature channel together".](readme-pic.jpg)

Setup details are below, but to run this you'll need a server running somewhere, like a VPS (I use [DigitalOcean](https://m.do.co/c/ce4b7db8e0a5); that's a referral link). My setup is configured to post to both Micro.blog and Twitter, with error monitoring by [Honeybadger](https://app.honeybadger.io/) (free plan), but those are all optional.

\* _A note about Twitter: it wasn't great when I made the first version of this app in 2021, but Micro.blog included the easy cross-posting to Twitter that it still offers to several social media platforms (e.g., Bluesky and Mastodon). Then Elon Musk bumbled along, and one of his genius moves was to make that API expensive to use, forcing Micro.blog to remove their cross-posting. Yet Twitter was where many of the original users and audience still were, so I held my nose and added that functionality directly to this server. (And it was a pain in the ass, because their APIs suck. Yes, _APIs_, plural: I had to use both the "deprecated" v1 to post media, and the newer v2 to post tweets.) As of this writing in late Nov. 2023, Musk's mismanagement, stupidity, and general disgustingness continues to drive the platform into the ground with ever-increasing speed._

## application configuration

The server is easy to run locally (at least it is on my Mac), where you can experiment with it and work out configuration issues before you commit to setting it up to run 24/7 somewhere on the internet: `go run txt2mary`. When run like this, you'll probably want to post test messages via `curl` or the like.

_Almost_ all the configuration is in the following files, with the exception of two environment variables needed by one of the Twitter libraries: `GOTWI_API_KEY` and `GOTWI_API_KEY_SECRET`. If you're configuring posts to that service, you'll need to set values for those, in addition to putting them in the config file below. What can I say, Twitter sucks. But the Go libraries I found made it way easier to deal with, and this is how one of them works.

### 1. create `config.json`

Copy the example configuration from [`fixtures/config_test.json`](fixtures/config_test.json) (also shown below) to the root directory and rename it `config.json`. Set the values as described here, or simply remove them if they don't apply. Unfortunately you can't "comment out" JSON, but if you rename the attributes – for example, changing `"Twitter"` to `"Twitter.nope"` – they'll be ignored.

```json
{
  "Logfile": "txt2mary.log",
  "Server": ":8088",
  "ServerRoute": "/txt",
  "UsersFilename": "./fixtures/users_test.json",
  "HoneybadgerAPIKey": "hbp_key",
  "MicroBlog": {
    "Token": "foo-bar-42",
    "Destination": "https://foo.micro.blog/",
    "TestDestination": "https://foo-test.micro.blog/"
  },
  "Twitter": {
    "ConsumerKey": "key123",
    "ConsumerSecret": "secret456",
    "AccessToken": "token789",
    "AccessTokenSecret": "secretabc",
    "TestAccount": false
  }
}
```

Configuration options:

- `Logfile` - the filename for logging; remove or set to `"stderr"` to see log messages in the console
- `Server` & `ServerRoute` - these determine the webserver port and path: the sample config shown when run locally would make the server listen on `http://localhost:8888/txt`. I leave the host (before the `:`) blank both here and on my VPS, and configured Twilio (see below) using my VPS' IP address, but you could set a registered domain here instead.
- `UsersFilename` - the filename for the allowlist and user naming you also need to set up (see below)
- `HoneybadgerAPIKey` - to enable optional error reporting to Honeybadger, enter your API key here
- `MicroBlog` - configuration needed to post to this social network
  - `Token` - your Micro.blog API token, from [this account page](https://micro.blog/account/apps)
  - `Destination` - the URL of your Micro.blog site
  - `TestDestination` - Micro.blog allows the (free) creation of a test blog in your account. If you want to use that for test posts (see below), this is where you configure it
- `Twitter` - configuration needed to post to this social network
  - `ConsumerKey`, `ConsumerSecret`, `AccessToken`, & `AccessTokenSecret` - all the API token junk you'll need from a Twitter developer account to allow direct posting to that site
  - `TestAccount` - an optional boolean; when `true`, the server will send test posts (see below) to this Twitter account

Changes to this file require a server restart to pick up.

### 2. create `users.json` 

Copy this example file from [`fixtures/users_test.json`](fixtures/users_test.json) (also shown below) to the root directory and rename it `users.json` (or copy it to wherever and name it whatever you want; that path and name are configurable via `UsersFilename`, see above).

```json
{
  "+15125551212": "Gon",
  "+15125551213": "Killua"
}
```

This mapping of phone numbers to names is also the allowlist for the app. It lists the phone numbers and names of the people who you have granted permission to send texts to this system.

With this example data, when a text is sent via Twilio from (512) 555-1212, it will be posted to Micro.blog and/or Twitter, attributed to "Gon". If a text is sent from any number other than these two, nothing will be posted, and the sender will receive a "you're not allowed to text here" message.

Changes to this file require a server restart to pick up.

## testing

You can test your configuration without sending repeated messages to your main, "production" Micro.blog or Twitter. Any message sent that begins with the string **"TEST: "** (including the space) will be treated as a test message, and only sent to test-enabled services.

For Micro.blog, you'll first create a test blog within that service's UI. Enter the URL assigned, typically `[your-username]-test.micro.blog`, as the `MicroBlog.TestDestination` value in `config.json` (see above).

To have a separate test-enabled Twitter, you'll need another account, and you'll need to go through the same developer enrollment you will for the "real" account. Once you do that, you can enter all the API key info for that service, plus set `Twitter.TestAccount` to `true`. The server will only read the first `Twitter` configuration, so if you have the real one in there, too, you'll want to rename it so it's ignored (in other words, you might have two blocks, one for `"Twitter - ignore this one"`, and the other for `"Twitter"`). Not the most elegant, but if you're testing at all, it's probably just during initial setup.

## twilio setup

You'll sign up for a Twilio account and register a phone number to receive texts. I did this years ago and don't remember much about the ins and outs, but it wasn't that hard to figure out. The key thing is in the "Messaging Configuration" section, you'll configure it so that when "A message comes in" it will **Webhook** to the **URL** where your server is running (including the port number and path you've configured), via **HTTP POST**. 

Twilio has a ["Programmable Messaging Logs" section](https://console.twilio.com/us1/monitor/logs/sms) where you can see the messages received and details about how they were handled.

The cost per message is low, twenty to thirty cents per message, though now the "A2P 10DLC registration" required for Twilio to _respond_ to your texters can add to the cost. That requires a couple of one-time charges (about $20), and an ongoing $2/month. Thanks, spammers! You can also skip this registration entirely, but your texters won't get the "message posted at this URL" reply.

## server setup

I'm not going to try to give a lot of direction, but below are some rough notes from my own setup on my Ubuntu VPS. In addition to the compiled binary, you of course need to upload your `config.json` and `users.json` files. 

1. create service - run `systemctl edit --force --full txt2mary.service`

```
[Unit]
Description=Txt2Mary
Documentation=https://github.com/cgrayson/txt2mary
After=network.target syslog.target
Wants=httpd.service

[Service]
Type=simple
WorkingDirectory=/var/www/txt2mary
ExecStart=/var/www/txt2mary/txt2mary
StandardOutput=file:/var/www/txt2mary/txt2mary.out
StandardError=inherit
Environment="GOTWI_API_KEY=foo"
Environment="GOTWI_API_KEY_SECRET=bar"

[Install]
WantedBy=multi-user.target
```

2. start the service: `systemctl start txt2mary.service` (and stop or restart it by replacing `start` with `stop` or `restart`).

## license

This software is licensed under the GNU General Public License v3.0. See [`COPYING`](COPYING).