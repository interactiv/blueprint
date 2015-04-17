// Copyright 2015 aikah
// License MIT

package backup

import (
	"archive/zip"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

/***********************************************/
/*                                             */
/*           Zipping operations                */
/*                                             */
/***********************************************/

type Archiver interface {
	Archive(src, dest string) error
	GetExtension() string
}

type zipper struct {
}

// we cast nil to *zipper type.
// since zipper has no fields.
// there will be no memory allocation
var ZIP Archiver = (*zipper)(nil)

func (z *zipper) GetExtension() string {
	return "zip"
}
func (z *zipper) Archive(src, dest string) error {
	var (
		err error
		out *os.File
		w   *zip.Writer
	)
	if err = os.MkdirAll(filepath.Dir(dest), 0777); err != nil {
		return err
	}
	if out, err = os.Create(dest); err != nil {
		return err
	}
	defer out.Close()
	w = zip.NewWriter(out)
	defer w.Close()
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil //skip
		}
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		f, err := w.Create(path)
		if err != nil {
			return err
		}
		io.Copy(f, in)
		return nil
	})
}

/***********************************************/
/*                                             */
/*             Hashing operations              */
/*                                             */
/***********************************************/

func DirHash(path string) (string, error) {
	hash := md5.New()
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		io.WriteString(hash, path)
		fmt.Fprintf(hash, "%v", info.IsDir())
		fmt.Fprintf(hash, "%v", info.ModTime())
		fmt.Fprintf(hash, "%v", info.Mode())
		fmt.Fprintf(hash, "%v", info.Size())
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

/***********************************************/
/*                                             */
/*           Monitoring operations             */
/*                                             */
/***********************************************/

type Monitor struct {
	Paths       map[string]string
	Archiver    Archiver
	Destination string
}

// Now detects changes in a hash of paths
func (m *Monitor) Now() (int, error) {
	var (
		counter int
	)
	for path, lastHash := range m.Paths {
		newHash, err := DirHash(path)
		if err != nil {
			return 0, err
		}
		if newHash != lastHash {
			err := m.act(path)
			if err != nil {
				return counter, err
			}
			m.Paths[path] = newHash // update hash
			counter++
		}

	}
	return counter, nil
}

func (m *Monitor) act(path string) error {
	dirname := filepath.Base(path)
	filename := fmt.Sprintf("%d.%s", time.Now().UnixNano(), m.Archiver.GetExtension())
	return m.Archiver.Archive(path, filepath.Join(m.Destination, dirname, filename))
}
