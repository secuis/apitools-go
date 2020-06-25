package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const (
	INFO  = 1
	ALERT = 2
)

type ImageAccessory struct {
	Type     string `json:"type"`
	ImageUrl string `json:"image_url"`
	AltText  string `json:"alt_text"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Section struct {
	Type      string          `json:"type"`
	Text      *Text           `json:"text,omitempty"`
	Accessory *ImageAccessory `json:"accessory,omitempty"` // This is a pointer because the encoder won't omit it otherwise
}

type Body struct {
	Blocks []Section `json:"blocks"`
}

type Slack struct {
	infohook, alerthook, environment string
	testMode                         bool
}

type ErrorWriter struct {
	slacker *Slack
	channel io.Writer
}

func (s ErrorWriter) Write(data []byte) (n int, err error) {
	re := regexp.MustCompile(`(?m)panic recovered:\n(.+)[\s\S]+\n\n(.+)`)
	msg := string(data)
	matches := re.FindAllStringSubmatchIndex(msg, -1)[0]

	// Formatting the error message
	fmtMsg := "*" + msg[matches[2]:matches[3]-1] + "*\n" + msg[matches[3]:matches[5]] +
		"\n```" + msg[matches[5]:] + "```"

	_ = s.slacker.Alert("@here ERROR", fmtMsg)
	_, _ = fmt.Fprint(s.channel, "\n\x1b[31m"+msg)
	return len(msg), nil
}

func NewSlack(infohook, alerthook, environment string, testMode bool) *Slack {
	return &Slack{
		infohook:    infohook,
		alerthook:   alerthook,
		environment: environment,
		testMode:    testMode,
	}
}

func (s *Slack) Writer(channel io.Writer) *ErrorWriter {
	return &ErrorWriter{slacker: s, channel: channel}
}

func (s *Slack) Info(headline, msg string) error {
	var body Body
	if s.testMode {
		body = formatTest(headline, msg, s.environment, INFO)
	} else {
		body = format(headline, msg, s.environment, INFO)
	}
	return s.send(body, s.infohook)
}

func (s *Slack) Alert(headline, msg string) error {
	var body Body
	if s.testMode {
		body = formatTest(headline, msg, s.environment, ALERT)
	} else {
		body = format(headline, msg, s.environment, ALERT)
	}
	return s.send(body, s.alerthook)
}

func (s *Slack) send(body Body, hook string) error {
	marshal, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	_, err = http.Post(hook, "application/json", bytes.NewBuffer(marshal))
	if err != nil {
		return err
	}
	return nil
}

func formatTest(headline, msg, environment string, level int) Body {
	body := format(headline, msg, environment, level)

	body.Blocks = append(body.Blocks, []Section{
		{
			Type: "divider",
		},
		{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: "*THE ABOVE IS A TEST, DON'T MIND IT*",
			},
			Accessory: &ImageAccessory{
				Type:     "image",
				ImageUrl: "http://icons.iconarchive.com/icons/streamlineicons/streamline-ux-free/1024/hacker-icon.png",
				AltText:  "testing",
			},
		},
	}...)

	return body
}

func format(headline, msg, environment string, level int) Body {
	var image *ImageAccessory
	switch level {
	case ALERT:
		image = &ImageAccessory{
			Type:     "image",
			ImageUrl: "https://www.freeiconspng.com/uploads/message-alert-red-icon--message-types-icons--softiconsm-4.png",
			AltText:  "alert",
		}
	case INFO:
		image = &ImageAccessory{
			Type:     "image",
			ImageUrl: "https://www.freeiconspng.com/uploads/light-bulb-icon---colorful-stickers-part-2-set-yellow-3.png",
			AltText:  "data",
		}
	}

	return Body{
		Blocks: []Section{
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: headline,
				},
			},
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: msg,
				},
				Accessory: image,
			},
			{
				Type: "divider",
			},
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: fmt.Sprintf("Environment: *%s*", environment),
				},
			},
		},
	}
}
