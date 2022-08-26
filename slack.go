package main

import (
	"github.com/slack-go/slack"
	_ "github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const gecChannel = "C03TUFXP6BG"

type socketmodeClient interface {
	Ack(socketmode.Request, ...interface{})
	Run() error
}

type slackClient interface {
	PostMessage(string, ...slack.MsgOption) (string, string, error)
	GetConversations(*slack.GetConversationsParameters) ([]slack.Channel, string, error)
	CreateConversation(string, bool) (*slack.Channel, error)
}

type Slack struct {
	s     socketmodeClient
	slack slackClient
}

func NewSlack(slackAppToken, slackBotToken string) (b Slack, err error) {
	b.slack = slack.New(slackBotToken,
		slack.OptionAppLevelToken(slackAppToken),
	)

	b.s = socketmode.New(
		b.slack.(*slack.Client),
	)

	go b.events()

	return
}

// Given a channel we're in, for every message from a non-bot
// we should enqueue the message back onto the user.
//
//
// Do this by writing to an outgoing stream which the gec-bot
// will pick up and respond with
func (s Slack) events() {
	s.s.Run()
	for range s.s.(*socketmode.Client).Events {
	}
}

func (s Slack) Send(m Message) (err error) {
	// get groups
	id, err := s.chanID(m.ID)
	if err != nil {
		return
	}

	if id == "" {
		id, err = s.newChannel(m.ID)
		if err != nil {
			return
		}

		// Ignore failures to this for now
		s.slack.PostMessage(gecChannel, slack.MsgOptionCompose(
			slack.MsgOptionText("New repsondent: #"+m.ID, false),
			slack.MsgOptionParse(true)),
		)
	} else {
		s.slack.PostMessage(gecChannel, slack.MsgOptionCompose(
			slack.MsgOptionText("Update from respondent: #"+m.ID, false),
			slack.MsgOptionParse(true)),
		)
	}

	// send message to group
	_, _, err = s.slack.PostMessage(id, slack.MsgOptionText(m.Message, false))

	return
}

func (s Slack) chanID(user string) (id string, err error) {
	channels, _, err := s.slack.GetConversations(&slack.GetConversationsParameters{Limit: 100})
	if err != nil {
		return
	}

	for _, channel := range channels {
		if channel.Name == user {
			return channel.ID, nil
		}
	}

	return
}

func (s Slack) newChannel(user string) (id string, err error) {
	c, err := s.slack.CreateConversation(user, false)
	if err != nil {
		return
	}

	id = c.ID

	if err != nil {
		panic(err)
	}

	return
}
