// Copyright 2014 The Cactus Authors. All rights reserved.

// +build ignore

package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type Cubed struct {
}

type CubedCreateArgs struct {
	Name string
	Data []byte
}

func (d *Cubed) Create(args *CubedCreateArgs, reply *int64) error {
	f, err := os.Create(args.Name)
	if err != nil {
		return err
	}
	*reply, err = io.Copy(f, bytes.NewReader(args.Data))
	if err != nil {
		return err
	}
	return f.Close()
}

type CubedExecuteArgs struct {
	Name string
	Args []string

	Stdin []byte

	LimitCpu    float64
	LimitMemory int
}

type CubedExecuteReply struct {
	Stdout []byte
	Stderr []byte

	UsageCpu    float64
	UsageMemory int

	Success bool
}

func (d *Cubed) Execute(args *CubedExecuteArgs, reply *CubedExecuteReply) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	path, err := exec.LookPath(args.Name)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := io.Copy(stdinW, bytes.NewReader(args.Stdin))
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

	proc, err := os.StartProcess(path, append([]string{args.Name}, args.Args...), &os.ProcAttr{
		Files: files,
	})
	if err != nil {
		return err
	}

	_, _, errno := syscall.RawSyscall6(syscall.SYS_PRLIMIT64, uintptr(proc.Pid), uintptr(syscall.RLIMIT_CPU), uintptr(unsafe.Pointer(&syscall.Rlimit{
		Cur: uint64(args.LimitCpu + 2),
		Max: uint64(args.LimitCpu + 2),
	})), 0, 0, 0)
	if errno != 0 {
		return errno
	}
	_, _, errno = syscall.RawSyscall6(syscall.SYS_PRLIMIT64, uintptr(proc.Pid), uintptr(syscall.RLIMIT_AS), uintptr(unsafe.Pointer(&syscall.Rlimit{
		Cur: uint64(args.LimitMemory+32) * (1 << 20),
		Max: uint64(args.LimitMemory+32) * (1 << 20),
	})), 0, 0, 0)
	if errno != 0 {
		return errno
	}

	state, err := proc.Wait()
	if err != nil {
		return err
	}

	for _, f := range files {
		f.Close()
	}
	wg.Wait()

	*reply = CubedExecuteReply{
		Stdout:      stdoutBuf.Bytes(),
		Stderr:      stderrBuf.Bytes(),
		UsageCpu:    (state.SystemTime() + state.UserTime()).Seconds(),
		UsageMemory: int(state.SysUsage().(*syscall.Rusage).Maxrss / (1 << 10)),
		Success:     state.Success(),
	}
	return nil
}

func (d *Cubed) Dispose(args int, reply *int) error {
	go func() {
		time.Sleep(512 * time.Microsecond)
		os.Exit(0)
	}()
	return nil
}

func main() {
	err := syscall.Setresuid(1000, 1000, 1000)
	catch(err)
	err = syscall.Setresuid(1000, 1000, 1000)
	catch(err)

	err = os.Chdir("/home/cactus")
	catch(err)

	rpc.Register(&Cubed{})
	rpc.HandleHTTP()
	err = http.ListenAndServe(":31337", nil)
	catch(err)
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
