package backup_test

import (
	backup "../backup"
	"github.com/interactiv/expect"
	"os"
	"path"
	"testing"
)

func TestArchive(t *testing.T) {
	e := expect.New(t)
	workingDir, err := os.Getwd()
	destPath := path.Join(workingDir, "test.zip")
	backup.ZIP.Archive(path.Join(workingDir, "test"), destPath)
	defer func() {
		err = os.Remove(destPath)
		e.Expect(err).ToBeNil()
	}()
	_, err = os.Stat(destPath)
	e.Expect(err).ToBeNil()
}

func TestDirHash(t *testing.T) {
	e := expect.New(t)
	// given an existing directory,returns a hash string
	wd, err := os.Getwd()
	e.Expect(err).ToBeNil()
	testPath := path.Join(wd, "test")
	res, err := backup.DirHash(testPath)
	e.Expect(err).ToBe(nil)
	t.Log(res)
}
