package main

import (
	"os"
)

var (
	redisAddr = os.Getenv("REDIS_ADDR")
	stream    = os.Getenv("INCOMING_STREAM")

	appToken = os.Getenv("APP_TOKEN")
	botToken = os.Getenv("BOT_TOKEN")
)

func main() {
	r, err := NewRedis(redisAddr, stream)
	if err != nil {
		panic(err)
	}

	s, err := NewSlack(appToken, botToken)
	if err != nil {
		panic(err)
	}

	m := make(chan Message)
	go func() {
		err := r.Process(m)
		if err != nil {
			panic(err)
		}
	}()

	for message := range m {
		err = s.Send(message)
		if err != nil {
			panic(err)
		}
	}
}
