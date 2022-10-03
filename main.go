package main

import (
	"os"

	"github.com/gender-equality-community/types"
)

var (
	redisAddr = os.Getenv("REDIS_ADDR")
	inStream  = os.Getenv("INCOMING_STREAM")
	outStream = os.Getenv("OUTGOING_STREAM")

	appToken = os.Getenv("APP_TOKEN")
	botToken = os.Getenv("BOT_TOKEN")

	m = make(chan types.Message)
)

func main() {
	r, err := NewRedis(redisAddr, inStream, outStream)
	if err != nil {
		panic(err)
	}

	s, err := NewSlack(appToken, botToken, r)
	if err != nil {
		panic(err)
	}

	go func() {
		err := r.Process(m)
		if err != nil {
			panic(err)
		}
	}()

	panic(messageLoop(s, m))
}

func messageLoop(s Slack, m chan types.Message) (err error) {
	for message := range m {
		if message.Message == "" {
			continue
		}

		err = s.Send(message)
		if err != nil {
			break
		}
	}

	return
}
