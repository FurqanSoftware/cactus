// Copyright 2014 The Cactus Authors. All rights reserved.

package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"runtime"

	"github.com/hjr265/go-zrsc/zrsc"
	"github.com/pelletier/go-toml"

	"github.com/hjr265/cactus/belt"
	"github.com/hjr265/cactus/cube"
	_ "github.com/hjr265/cactus/rpc"
)

var cfg *toml.TomlTree

func main() {
	runtime.GOMAXPROCS((runtime.NumCPU() + 1) / 2)

	configName := flag.String("c", "config.tml", "")
	flag.Parse()

	_, err := os.Stat("config.tml")
	if os.IsNotExist(err) {
		f2, err := zrsc.Open("cmd/cactus/config-sample.tml")
		catch(err)

		f, err := os.Create("config.tml")
		_, err = io.Copy(f, f2)
		catch(err)

		err = f2.Close()
		catch(err)
		err = f.Close()
		catch(err)
	}

	cfg, err = toml.LoadFile(*configName)
	catch(err)

	if flag.Arg(0) != "belt" {
		rpc.HandleHTTP()

		go func() {
			addr, ok := cfg.Get("core.addr").(string)
			if !ok {
				log.Fatal("Missing core.addr in config.tml")
			}

			log.Printf("Listening on %s", addr)
			err := http.ListenAndServe(addr, nil)
			catch(err)
		}()

		belt.Queue = belt.NewLocalQueue()

	} else {
		belt.Queue, err = belt.NewRemoteQueue(flag.Arg(1))
		catch(err)
	}

	if cube.DockerExe != "" {
		log.Print("Docker detected")

		image, ok := cfg.Get("cube.docker.image").(string)
		if !ok {
			log.Fatal("Missing cube.docker.image in config.tml")
		}
		cube.DockerImage = image
	}

	beltSize, ok := cfg.Get("belt.size").(int64)
	if !ok {
		log.Fatal("Missing belt.size in config.tml")
	}
	for ; beltSize > 0; beltSize-- {
		go belt.Loop()
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)

	log.Printf("Received %s; exiting", <-sigCh)
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}

func trace(err error) {
	if err != nil {
		log.Print(err)
	}
}
