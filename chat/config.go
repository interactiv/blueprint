package main

import (
	"bytes"

	"encoding/json"
	"log"
	"os"
)

type config struct {
	SecurityKey string
	Github      struct {
		ClientId string
		Secret   string
		Callback string
	}
	Google struct {
		ClientId string
		Secret   string
		Callback string
	}
}

func NewConfigFromFilePath(configPath string) *config {
	var (
		c    *config = new(config)
		err  error
		file *os.File
	)
	if file, err = os.Open(configPath); err != nil {
		log.Fatal(err)
	}
	if err = json.NewDecoder(file).Decode(c); err != nil {
		log.Fatal(err)
	}
	return c
}

func NewConfigFromString(configString string) *config {
	var (
		c   *config = new(config)
		err error
	)
	if err = json.NewDecoder(bytes.NewBufferString(configString)).Decode(c); err != nil {
		log.Fatal(err)
	}

	return c
}
