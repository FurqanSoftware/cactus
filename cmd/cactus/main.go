// Copyright 2014 The Cactus Authors. All rights reserved.

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	"github.com/FurqanSoftware/cactus/api"
	"github.com/FurqanSoftware/cactus/belt"
	"github.com/FurqanSoftware/cactus/data"
	"github.com/FurqanSoftware/cactus/sandbox"
)

var (
	version   string
	date      string
	repoOwner = "FurqanSoftware"
	repoName  = "cactus"
)

func main() {
	printBanner()

	runtime.GOMAXPROCS((runtime.NumCPU() + 1) / 2)

	_, err := os.Stat("config.toml")
	if os.IsNotExist(err) {
		log.Print("Creating config.toml from sample")

		f2, err := configSampleFS.Open("config-sample.toml")
		catch(err)

		f, err := os.Create("config.toml")
		catch(err)
		_, err = io.Copy(f, f2)
		catch(err)

		err = f2.Close()
		catch(err)
		err = f.Close()
		catch(err)
	}

	cfg, err := parseConfig("config.toml")
	catch(err)
	log.Print("Loaded config.toml")

	if cfg.Core.Addr == "" {
		log.Fatal("Missing core.addr in config.toml")
	}

	dbPath := cfg.Data.DB
	if dbPath == "" {
		dbPath = "cactus.db"
	}
	blobsPath := cfg.Data.Blobs
	if blobsPath == "" {
		blobsPath = "cactus.bs"
	}
	log.Printf("Opening database %s and blob store %s", dbPath, blobsPath)
	err = data.Open(dbPath, blobsPath)
	catch(err)

	err = api.InitStore()
	catch(err)

	go func() {
		log.Printf("Listening on %s", cfg.Core.Addr)
		err := http.ListenAndServe(cfg.Core.Addr, nil)
		catch(err)
	}()

	sandboxMode := cfg.Sandbox.Mode
	if sandboxMode == "" {
		sandboxMode = "jail"
	}
	switch sandboxMode {
	case "jail":
		belt.NewCell = func() (sandbox.Cell, error) {
			return sandbox.NewJailCell()
		}
	case "docker":
		belt.NewCell = func() (sandbox.Cell, error) {
			return sandbox.NewDockerCell()
		}
		if cfg.Sandbox.DockerImage != "" {
			sandbox.DockerImage = cfg.Sandbox.DockerImage
		}
	default:
		log.Fatalf("Unknown sandbox.mode value %q in config.toml (expected \"jail\" or \"docker\")", sandboxMode)
	}
	log.Printf("Sandbox mode: %s", sandboxMode)

	if cfg.Belt.Size == 0 {
		log.Fatal("Missing belt.size in config.toml")
	}
	log.Printf("Starting %d judging worker(s)", cfg.Belt.Size)
	for i := 0; i < cfg.Belt.Size; i++ {
		go belt.Loop()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	log.Printf("Received %s; exiting", <-sigCh)
}

func printBanner() {
	w := log.Writer()
	fmt.Fprintln(w, `   ____           _             `)
	fmt.Fprintln(w, "  / ___|__ _  ___| |_ _   _ ___ ")
	fmt.Fprintln(w, " | |   / _` |/ __| __| | | / __|")
	fmt.Fprintln(w, ` | |__| (_| | (__| |_| |_| \__ \`)
	fmt.Fprintln(w, `  \____\__,_|\___|\__|\__,_|___/`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "By Furqan Software (https://furqansoftware.com)")
	fmt.Fprintln(w)

	if version != "" {
		fmt.Fprintf(w, "» Release: %s", version)
	} else {
		fmt.Fprint(w, "» Release: -")
	}
	if date != "" {
		fmt.Fprintf(w, " (%s)", date)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	fmt.Fprintf(w, "» Project: https://github.com/%s/%s\n", repoOwner, repoName)
	fmt.Fprintln(w)
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
