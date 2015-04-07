package main

import (
	"github.com/interactiv/expect"
	"testing"
)

func TestNewConfigFromFilePath(t *testing.T) {
	config := NewConfigFromFilePath("./config.json")
	expect.Expect(config.SecurityKey, t).Not().ToBeNil()
}

func TestNewConfigFromString(t *testing.T) {
	config := NewConfigFromString(`{"securityKey":"foo"}`)
	expect.Expect(config.SecurityKey, t).ToBe("foo")
}
