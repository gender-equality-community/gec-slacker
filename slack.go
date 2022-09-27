package main

import (
	"fmt"
	"regexp"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const gecChannel = "C03TUFXP6BG"

var joinerMsgRe = regexp.MustCompile(`.+ has joined the channel`)

type socketmodeClient interface {
	Ack(socketmode.Request, ...interface{})
	Run() error
}

type slackClient interface {
	PostMessage(string, ...slack.MsgOption) (string, string, error)
	GetConversations(*slack.GetConversationsParameters) ([]slack.Channel, string, error)
	CreateConversation(string, bool) (*slack.Channel, error)
	JoinConversation(string) (*slack.Channel, string, []string, error)
	GetConversationInfo(string, bool) (*slack.Channel, error)
}

type Slack struct {
	s     socketmodeClient
	slack slackClient
	p     *Redis
}

func NewSlack(slackAppToken, slackBotToken string, p Redis) (b Slack, err error) {
	b.p = &p
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
// Do this by writing to an outgoing stream which the gec-bot
// will pick up and respond with
func (s Slack) events() {
	var err error

	go s.s.Run()
	for evt := range s.s.(*socketmode.Client).Events {
		err = s.handleEvent(evt)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s Slack) handleEvent(evt socketmode.Event) (err error) {
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		eventsAPIEvent, _ := evt.Data.(slackevents.EventsAPIEvent)

		switch eventsAPIEvent.Type {
		case slackevents.CallbackEvent:
			innerEvent := eventsAPIEvent.InnerEvent

			switch ev := innerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				err = s.handleMessageEvent(ev)
				if err != nil {
					return
				}
			}
		}

		if evt.Request != nil {
			s.s.Ack(*evt.Request)
		}
	}

	return
}

func (s Slack) handleMessageEvent(ev *slackevents.MessageEvent) (err error) {
	// Don't forward anything in a thread; allow people to make
	// comments instead
	if ev.ThreadTimeStamp != "" {
		return
	}

	// Ignore anything from a bot
	if ev.BotID != "" {
		return
	}

	// Ignore messages where a user has joined a channel
	if joinerMsgRe.Match([]byte(ev.Text)) {
		return
	}

	msg := ev.Text

	// skip messages where body is empty (such as when a message is
	// deleted)
	if msg == "" {
		return
	}

	channel, err := s.chanName(ev.Channel)
	if err != nil {
		return
	}

	return s.p.Produce(channel, msg)
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
		_, _, err = s.slack.PostMessage(gecChannel, slack.MsgOptionCompose(
			slack.MsgOptionText("New repsondent: #"+m.ID, false),
			slack.MsgOptionParse(true)),
		)
	} else {
		_, _, err = s.slack.PostMessage(gecChannel, slack.MsgOptionCompose(
			slack.MsgOptionText("Update from respondent: #"+m.ID, false),
			slack.MsgOptionParse(true)),
		)
	}

	if err != nil {
		return
	}

	// send message to group
	_, _, err = s.slack.PostMessage(id, slack.MsgOptionText(m.Message, false))

	return
}

func (s Slack) chanName(id string) (name string, err error) {
	c, err := s.slack.GetConversationInfo(id, false)
	if err != nil {
		return
	}

	name = c.Name

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

	_, _, _, err = s.slack.JoinConversation(id)

	return
}
