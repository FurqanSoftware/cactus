package zrsc

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	initErr error

	bs = map[string][]byte{}
	fs = map[string]*zip.File{}
	ds = map[string][]os.FileInfo{}
)

type HttpDir string

func (d HttpDir) Open(name string) (http.File, error) {
	return Open(strings.TrimLeft(path.Clean(path.Join(string(d), name)), "/"))
}

type File struct {
	*bytes.Reader

	name string
}

func (f *File) Close() error {
	return nil
}

func (f *File) Stat() (os.FileInfo, error) {
	return fs[f.name].FileInfo(), nil
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	d := ds[f.name]
	if count > 0 && len(d) > count {
		d = d[:count]
	}
	return d, nil
}

func Open(name string) (*File, error) {
	if initErr != nil {
		return nil, initErr
	}

	zf, ok := fs[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	rc, err := zf.Open()
	if err != nil {
		return nil, err
	}

	b, ok := bs[name]
	if !ok {
		b, err = ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		err = rc.Close()
		if err != nil {
			return nil, err
		}

		bs[name] = b
	}

	return &File{
		Reader: bytes.NewReader(b),
		name:   name,
	}, nil
}

func init() {
	defer func() {
		val := recover()
		if val == nil {
			return
		}
		if err, ok := val.(error); ok {
			initErr = err
		} else {
			panic(val)
		}
	}()

	name, err := filepath.Abs(os.Args[0])
	catch(err)

	f, err := os.Open(name)
	catch(err)
	fi, err := f.Stat()
	catch(err)

	r, err := zip.NewReader(f, fi.Size())
	catch(err)

	for _, f := range r.File {
		fi := f.FileInfo()
		if fi.IsDir() {
			ds[f.Name] = []os.FileInfo{}
		}

		fs[f.Name] = f

		dir := path.Dir(f.Name)
		ds[dir] = append(ds[dir], fi)
	}
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
