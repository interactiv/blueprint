// avatar_test.go
package main

import (
	"github.com/interactiv/expect"
	g "github.com/stretchr/gomniauth/test"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestAuthAvatar(t *testing.T) {
	var authAvatar AuthAvatar
	testUser := &g.TestUser{}
	testUser.On("AvatarURL").Return("", ErroNoAvatarURL)
	testChatUser := &chatUser{User: testUser}
	url, err := authAvatar.GetAvatarURL(testChatUser)
	expect.Expect(err, t).ToBe(ErroNoAvatarURL)
	testUrl := "http://url-to-avatar"
	testUser = &g.TestUser{}
	testChatUser.User = testUser
	testUser.On("AvatarURL").Return(testUrl, nil)
	url, err = authAvatar.GetAvatarURL(testChatUser)
	expect.Expect(err, t).ToBeNil()
	expect.Expect(url, t).ToBe(testUrl)
}

func TestGravatarAvatar(t *testing.T) {
	var gravatarAvatar GravatarAvatar
	e := expect.New(t)
	user := &chatUser{uniqueID: "abc"}
	url, err := gravatarAvatar.GetAvatarURL(user)
	e.Expect(err).ToBeNil()
	e.Expect(url).ToBe("//www.gravatar.com/avatar/abc")
}

func TestFileSystemAvatar(t *testing.T) {
	e := expect.New(t)
	//make a test file
	filename := path.Join("avatars", "abc.jpg")
	ioutil.WriteFile(filename, []byte{}, 0644)
	defer func() { os.Remove(filename) }()
	var fileSystemAvatar FileSystemAvatar
	user := &chatUser{uniqueID: "abc"}
	url, err := fileSystemAvatar.GetAvatarURL(user)
	e.Expect(err).ToBeNil()
	e.Expect(url).ToContain("abc")

}
