package main

type Message struct {
	ID      string `mapstructure:"id"`
	Ts      string `mapstructure:"ts"`
	Message string `mapstructure:"msg"`
}
