// Copyright 2014 The Cactus Authors. All rights reserved.

package cube

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Plain struct {
	*rpc.Client

	dir string
}

func (s *Plain) Init() error {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	s.dir = dir
	return nil
}

func (s *Plain) Create(name string, data io.Reader) error {
	f, err := os.Create(filepath.Join(s.dir, name))
	if err != nil {
		return err
	}
	_, err = io.Copy(f, data)
	if err != nil {
		return err
	}
	return f.Close()
}

func (s *Plain) Execute(proc *Process) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	_, err := os.Open(filepath.Join(s.dir, proc.Name))
	path := ""
	if os.IsNotExist(err) {
		path, err = exec.LookPath(proc.Name)
		if err != nil {
			return err
		}

	} else {
		path = proc.Name
	}

	wg := sync.WaitGroup{}

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := io.Copy(stdinW, bytes.NewReader(proc.Stdin))
		trace(err)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return err
	}
	stdoutBuf := &bytes.Buffer{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := io.Copy(stdoutBuf, stdoutR)
		trace(err)
		err = stdoutR.Close()
		trace(err)
	}()

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		return err
	}
	stderrBuf := &bytes.Buffer{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := io.Copy(stderrBuf, stderrR)
		trace(err)
		err = stderrR.Close()
		trace(err)
	}()

	files := []*os.File{
		stdinR,
		stdoutW,
		stderrW,
	}

	start := time.Now()
	proc2, err := os.StartProcess(path, append([]string{proc.Name}, proc.Args...), &os.ProcAttr{
		Dir:   s.dir,
		Files: files,
	})
	if err != nil {
		return err
	}

	chWait := make(chan *os.ProcessState)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		state, err := proc2.Wait()
		trace(err)

		for _, f := range files {
			f.Close()
		}

		chWait <- state
		close(chWait)
	}()

	select {
	case state := <-chWait:
		proc.Success = state.Success()

	case <-time.After(time.Duration(proc.Limits.Cpu+1) * time.Second):
		err = proc2.Kill()
		trace(err)
	}

	wg.Wait()

	proc.Stdout = stdoutBuf.Bytes()
	proc.Stderr = stderrBuf.Bytes()
	proc.Usages.Cpu = time.Now().Sub(start).Seconds()

	return nil
}

func (s *Plain) Dispose() error {
	// return os.RemoveAll(s.dir)
	return nil
}
