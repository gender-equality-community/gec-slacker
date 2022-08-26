package main

import (
	"log"
	"os"
)

var (
	redisAddr = os.Getenv("REDIS_ADDR")
	inStream  = os.Getenv("INCOMING_STREAM")
	outStream = os.Getenv("OUTGOING_STREAM")

	appToken = os.Getenv("APP_TOKEN")
	botToken = os.Getenv("BOT_TOKEN")
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

	m := make(chan Message)
	go func() {
		err := r.Process(m)
		if err != nil {
			panic(err)
		}
	}()

	for message := range m {
		log.Printf("%#v", message)

		err = s.Send(message)
		if err != nil {
			panic(err)
		}
	}
}
