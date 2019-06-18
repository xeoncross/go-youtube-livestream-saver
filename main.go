package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

func main() {

	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Watch %q\n", config.Channel)

	ctx, cancel := context.WithCancel(context.Background())

	// Abort when we press CTRL+C or send the kill signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for x := range c {
			log.Println(x.String(), "Signal. Shutdown started.")
			cancel()
		}
	}()

	watch(config, ctx)

	log.Println("Clean shutdown")
}

type Config struct {
	APIKey  string `json:"api_key"`
	Channel string `json:"channel"`
}

func loadConfig() (*Config, error) {
	c := &Config{}

	f, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func watch(config *Config, ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			err := PollLivestreams(config, ctx)
			if err != nil {
				return err
			}
		}
	}
}

func PollLivestreams(config *Config, ctx context.Context) error {
	yt, err := youtube.NewService(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return err
	}

	s := yt.Search.List("snippet")

	s.ChannelId(config.Channel)
	s.EventType("live")
	// s.MaxResults(25)
	s.Type("video")

	response, err := s.Do()

	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", response)

	if len(response.Items) == 0 {
		log.Println("No livestreams")
		return nil
	}

	for _, item := range response.Items {
		fmt.Printf("%v\n", item)
	}

	return nil
}
