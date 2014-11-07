// Copyright 2014 The Cactus Authors. All rights reserved.

package ui

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/hjr265/go-zrsc/zrsc"
)

func ServeAsset(w http.ResponseWriter, r *http.Request) {
	name := path.Join("ui", "assets", r.URL.Path)
	f, err := zrsc.Open(name)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, name)
		return
	}
	catch(err)
	defer f.Close()
	fi, err := f.Stat()
	catch(err)

	modSince, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since"))
	if err == nil && fi.ModTime().Before(modSince.Add(1*time.Second)) {
		w.Header().Set("Cache-Control", "public")
		http.Error(w, "", http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(fi.Name())))
	w.Header().Set("Cache-Control", "public")
	w.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
	_, err = io.Copy(w, f)
	catch(err)
}
