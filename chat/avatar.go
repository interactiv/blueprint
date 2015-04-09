// avatar.go
package main

import (
	"errors"
	"io/ioutil"
	"path"
)

var ErroNoAvatarURL = errors.New("chat: Unable to get an avatar URL.")

type Avatar interface {
	GetAvatarURL(c *Client) (string, error)
}

type AuthAvatar struct{}

var UseAuthAvatar AuthAvatar

func (_ AuthAvatar) GetAvatarURL(c *Client) (string, error) {
	if url, ok := c.userData["avatar_url"]; ok {
		if urlStr, ok := url.(string); ok {
			return urlStr, nil
		}
	}
	return "", ErroNoAvatarURL
}

type GravatarAvatar struct{}

var UseGravatar GravatarAvatar

func (_ GravatarAvatar) GetAvatarURL(c *Client) (string, error) {
	if userId, ok := c.userData["userId"]; ok {
		if userIdString, ok := userId.(string); ok {
			return "//www.gravatar.com/avatar/" + userIdString, nil
		}
	}
	return "", ErroNoAvatarURL
}

type FileSystemAvatar struct {
}

var UseFileSystemAvatar FileSystemAvatar

func (_ FileSystemAvatar) GetAvatarURL(c *Client) (string, error) {
	if userId, ok := c.userData["userId"]; ok {
		if userIdString, ok := userId.(string); ok {
			if files, err := ioutil.ReadDir("avatars"); err == nil {
				for _, file := range files {
					if file.IsDir() {
						continue
					}
					if match, _ := path.Match(userIdString+"*", file.Name()); match {
						return "/avatars/" + file.Name(), nil
					}
				}
			}
		}
	}
	return "", ErroNoAvatarURL
}
