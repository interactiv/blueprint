// avatar_test.go
package main

import (
	"github.com/interactiv/expect"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestAuthAvatar(t *testing.T) {
	var authAvatar AuthAvatar
	client := new(Client)
	url, err := authAvatar.GetAvatarURL(client)
	expect.Expect(err, t).ToEqual(ErroNoAvatarURL)
	testUrl := "http://url-to-avatar"
	client.userData = map[string]interface{}{
		"avatar_url": testUrl,
	}
	url, err = authAvatar.GetAvatarURL(client)
	expect.Expect(err, t).ToBeNil()
	expect.Expect(url, t).ToEqual(testUrl)
}

func TestGravatarAvatar(t *testing.T) {
	var gravatarAvatar GravatarAvatar
	e := expect.New(t)
	client := new(Client)
	client.userData = map[string]interface{}{
		"userId": "fd876f8cd6a58277fc664d47ea10ad19",
	}
	url, err := gravatarAvatar.GetAvatarURL(client)
	e.Expect(err).ToBeNil()
	e.Expect(url).ToBe("//www.gravatar.com/avatar/fd876f8cd6a58277fc664d47ea10ad19")
}

func TestFileSystemAvatar(t *testing.T) {
	e := expect.New(t)
	//make a test file
	filename := path.Join("avatars", "abc.jpg")
	ioutil.WriteFile(filename, []byte{}, 0644)
	defer func() { os.Remove(filename) }()
	var fileSystemAvatar FileSystemAvatar
	client := new(Client)
	client.userData = map[string]interface{}{
		"userId": "abc",
	}
	url, err := fileSystemAvatar.GetAvatarURL(client)
	e.Expect(err).ToBeNil()
	e.Expect(url).ToContain(client.userData["userId"].(string))

}
