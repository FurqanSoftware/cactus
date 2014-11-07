// Copyright 2014 The Cactus Authors. All rights reserved.

package ui

import (
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/hjr265/go-zrsc/zrsc"

	"github.com/hjr265/cactus/data"
)

var tplIndex = template.New("index.min.html")

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
	err := tplIndex.Execute(w, map[string]interface{}{
		"consts": map[string]interface{}{
			"Participant":         data.Participant,
			"Judge":               data.Judge,
			"Administrator":       data.Administrator,
			"Unresponded":         data.Unresponded,
			"Ignored":             data.Ignored,
			"Answered":            data.Answered,
			"Broadcasted":         data.Broadcasted,
			"Accepted":            data.Accepted,
			"WrongAnswer":         data.WrongAnswer,
			"CpuLimitExceeded":    data.CpuLimitExceeded,
			"MemoryLimitExceeded": data.MemoryLimitExceeded,
			"CompilationError":    data.CompilationError,
		},
	})
	catch(err)
}

func init() {
	f, err := zrsc.Open("ui/index.min.html")
	catch(err)
	b, err := ioutil.ReadAll(f)
	catch(err)
	_, err = tplIndex.Parse(string(b))
	catch(err)
	err = f.Close()
	catch(err)
}
