package main

import (
	"fmt"
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type dummySlackSMClient struct {
	wasAcked bool
	error    bool
}

func (d *dummySlackSMClient) Ack(socketmode.Request, ...interface{}) {
	d.wasAcked = true
}

func (d *dummySlackSMClient) Run() error {
	if d.error {
		return fmt.Errorf("an error")
	}

	return nil
}

type dummySlackClient struct {
	msg     string
	channel string
	error   bool
}

func (d *dummySlackClient) PostMessage(c string, _ ...slack.MsgOption) (string, string, error) {
	if d.error {
		return "", "", fmt.Errorf("an error")
	}

	d.channel = c

	return "", "", nil
}

func (d *dummySlackClient) GetConversations(*slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	if d.error {
		return nil, "", fmt.Errorf("an error")
	}

	return []slack.Channel{
		{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					ID: "foo",
				},
				Name: "foo",
			},
		},
	}, "", nil
}

func (d *dummySlackClient) CreateConversation(string, bool) (*slack.Channel, error) {
	if d.error {
		return nil, fmt.Errorf("an error")
	}

	return &slack.Channel{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID: "foobar",
			},
			Name: "foobar",
		},
	}, nil
}

func (d *dummySlackClient) JoinConversation(string) (*slack.Channel, string, []string, error) {
	if d.error {
		return nil, "", nil, fmt.Errorf("an error")
	}

	return nil, "", nil, nil
}

func (d *dummySlackClient) GetConversationInfo(string, bool) (*slack.Channel, error) {
	// same signature and data, be a little lazy
	return d.CreateConversation("", false)
}

func TestNewSlack(t *testing.T) {
	_, err := NewSlack("test-token", "test-bot-token", Redis{})
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
}

func TestSlack_handleEvent(t *testing.T) {
	for _, test := range []struct {
		name           string
		s              Slack
		evt            socketmode.Event
		expectChanName string
		expectMsg      string
		expectError    bool
		expectAck      bool
	}{
		{"not an EventTypeEventsAPI message", Slack{s: new(dummySlackSMClient), p: &Redis{client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeConnected}, "", "", false, false},
		{"not a CallbackEvent", Slack{s: new(dummySlackSMClient), p: &Redis{client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.AppRateLimited}, Request: new(socketmode.Request)}, "", "", false, true},
		{"not a MessageEvent", Slack{s: new(dummySlackSMClient), p: &Redis{client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.CallbackEvent, Data: new(slackevents.LinkSharedEvent)}, Request: new(socketmode.Request)}, "", "", false, true},
		{"message in a thread is skipped", Slack{s: new(dummySlackSMClient), p: &Redis{client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.CallbackEvent, InnerEvent: slackevents.EventsAPIInnerEvent{Data: &slackevents.MessageEvent{ThreadTimeStamp: "123456789"}}}, Request: new(socketmode.Request)}, "", "", false, true},
		{"message from a bot is skipped", Slack{s: new(dummySlackSMClient), p: &Redis{client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.CallbackEvent, InnerEvent: slackevents.EventsAPIInnerEvent{Data: &slackevents.MessageEvent{BotID: "123456789"}}}, Request: new(socketmode.Request)}, "", "", false, true},
		{"notification of a user joining is skipped", Slack{s: new(dummySlackSMClient), p: &Redis{client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.CallbackEvent, InnerEvent: slackevents.EventsAPIInnerEvent{Data: &slackevents.MessageEvent{Text: "so and so has joined the channel"}}}, Request: new(socketmode.Request)}, "", "", false, true},
		{"failure to determine channel message is sent on bombs out", Slack{s: new(dummySlackSMClient), slack: &dummySlackClient{error: true}, p: &Redis{client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.CallbackEvent, InnerEvent: slackevents.EventsAPIInnerEvent{Data: &slackevents.MessageEvent{Text: "hello, world!"}}}, Request: new(socketmode.Request)}, "", "", true, false},
		{"producer errors float up", Slack{s: new(dummySlackSMClient), slack: &dummySlackClient{}, p: &Redis{client: &dummyRedis{addError: true}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.CallbackEvent, InnerEvent: slackevents.EventsAPIInnerEvent{Data: &slackevents.MessageEvent{Text: "hello, world!"}}}, Request: new(socketmode.Request)}, "", "", true, false},
		{"slack messages correctly send to producer", Slack{s: new(dummySlackSMClient), slack: &dummySlackClient{}, p: &Redis{outStream: "test0", client: &dummyRedis{}}}, socketmode.Event{Type: socketmode.EventTypeEventsAPI, Data: slackevents.EventsAPIEvent{Type: slackevents.CallbackEvent, InnerEvent: slackevents.EventsAPIInnerEvent{Data: &slackevents.MessageEvent{Text: "hello, world!"}}}, Request: new(socketmode.Request)}, "test0", "hello, world!", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.s.handleEvent(test.evt)
			if err != nil && !test.expectError {
				t.Errorf("unexpected error: %v", err)
			} else if err == nil && test.expectError {
				t.Error("expected error")
			}

			dp := test.s.p.client.(*dummyRedis)

			t.Run("channel name", func(t *testing.T) {
				if test.expectChanName != dp.stream {
					t.Errorf("expected %q, received %q", test.expectChanName, dp.stream)
				}
			})

			t.Run("message body", func(t *testing.T) {
				var (
					got string
				)

				got, _ = dp.msg["msg"].(string)

				if test.expectMsg != got {
					t.Errorf("expected %q, received %q", test.expectMsg, got)
				}
			})

			t.Run("slack message was ack'd", func(t *testing.T) {
				got := test.s.s.(*dummySlackSMClient).wasAcked

				if test.expectAck != got {
					t.Errorf("expected %v, received %v", test.expectAck, got)
				}
			})
		})
	}
}
