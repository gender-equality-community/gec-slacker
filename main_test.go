package main

import (
	"testing"
	"time"
)

func TestMessageLoop(t *testing.T) {
	for _, test := range []struct {
		name        string
		input       Message
		slack       Slack
		expectChan  string
		expectError bool
	}{
		{"Empty messages are skipped", Message{}, Slack{slack: &dummySlackClient{}}, "", false},
		{"Slack errors bubble up", Message{Message: "hi"}, Slack{slack: &dummySlackClient{error: true}}, "", true},
		{"Valid messages from a known respondent are sent to slack", Message{Message: "hi", ID: "foo", Ts: "0"}, Slack{slack: &dummySlackClient{}, p: &Redis{client: &dummyRedis{}}}, "foo", false},
		{"Valid messages from an unknown respondent are sent to slack", Message{Message: "hi", ID: "xxx", Ts: "0"}, Slack{slack: &dummySlackClient{}, p: &Redis{client: &dummyRedis{}}}, "foobar", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := make(chan Message)

			go func() {
				m <- test.input
				time.Sleep(100 * time.Millisecond)
				close(m)
			}()

			err := messageLoop(test.slack, m)
			if err != nil && !test.expectError {
				t.Errorf("unexpected error: %v", err)
			} else if err == nil && test.expectError {
				t.Error("expected error")
			}

			// We have to test the slack channel's name/ ID, rather than
			// the body of the message becasue the Slack library doesn't
			// give you anyway of seeing what that message was when writing unit
			// tests because the moron who wrote it is a fucking idiot and holy shit
			// this so called official library reads like some fucking 'My First Go
			// Project' bull shit.
			//
			// Luckily, formatting messages to Slack is well tested in other places.
			t.Run("channel name", func(t *testing.T) {
				got := test.slack.slack.(*dummySlackClient).channel
				if test.expectChan != got {
					t.Errorf("expected %q, received %q", test.expectChan, got)
				}
			})
		})
	}
}
