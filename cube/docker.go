// Copyright 2014 The Cactus Authors. All rights reserved.

package cube

import (
	"io"
	"io/ioutil"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hjr265/go-zrsc/zrsc"
)

var (
	DockerExe = ""

	cubedHost = "localhost"
	cubedDir  = ""
)

type Docker struct {
	*rpc.Client

	cid string
}

func (s *Docker) Init() error {
	cmd := exec.Command("docker", "run", "--cap-add=SETUID", "--cap-add=SETGID", "--detach=true", "--expose=31337", "--interactive=true", "--publish=31337", "--volume="+cubedDir+":/root", "--workdir=/root", "--user=root", "tophws/cactus-cube", "go", "run", "cubed.go")
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	s.cid = strings.TrimSpace(string(out))

	cmd = exec.Command("docker", "port", s.cid)
	out, err = cmd.Output()
	if err != nil {
		return err
	}

	addr := strings.Split(strings.TrimSpace(string(out)), " -> ")[1]
	port := strings.Split(strings.TrimSpace(addr), ":")[1]

	for u := time.Now(); time.Now().Sub(u) < 8*time.Second; {
		s.Client, err = rpc.DialHTTP("tcp", cubedHost+":"+port)
		if err == nil {
			break
		}
		time.Sleep(64 * time.Microsecond)
	}
	return err
}

func (s *Docker) Create(name string, data io.Reader) error {
	dataBuf, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	return s.Call("Cubed.Create", &struct {
		Name string
		Data []byte
	}{
		Name: name,
		Data: dataBuf,
	}, new(int64))
}

func (s *Docker) Execute(proc *Process) error {
	reply := struct {
		Stdout      []byte
		Stderr      []byte
		UsageCpu    float64
		UsageMemory int
		Success     bool
	}{}
	err := s.Call("Cubed.Execute", &struct {
		Name        string
		Args        []string
		Stdin       []byte
		LimitCpu    float64
		LimitMemory int
	}{
		Name:        proc.Name,
		Args:        proc.Args,
		Stdin:       proc.Stdin,
		LimitCpu:    proc.Limits.Cpu,
		LimitMemory: proc.Limits.Memory,
	}, &reply)
	if err != nil {
		return err
	}
	proc.Stdout = reply.Stdout
	proc.Stderr = reply.Stderr
	proc.Usages.Cpu = reply.UsageCpu
	proc.Usages.Memory = reply.UsageMemory
	proc.Success = reply.Success
	return nil
}

func (s *Docker) Dispose() error {
	return nil

	err := s.Call("Cubed.Dispose", 0, new(int))
	if err != nil {
		return err
	}
	err = s.Client.Close()
	if err != nil {
		return err
	}

	exec.Command("docker", "kill", s.cid).Run()
	return exec.Command("docker", "rm", s.cid).Run()
}

func init() {
	cmd := exec.Command("docker", "images")
	err := cmd.Run()
	if err != nil {
		return
	}

	DockerExe = cmd.Path

	switch runtime.GOOS {
	case "linux":
		cubedDir, err = ioutil.TempDir("", "")
		catch(err)

	default:
		wd, err := os.Getwd()
		catch(err)
		cubedDir, err = ioutil.TempDir(wd, ".t-")
		catch(err)

		cmd := exec.Command("boot2docker", "ip")
		out, err := cmd.Output()
		catch(err)
		cubedHost = strings.TrimSpace(string(out))
	}

	f, err := zrsc.Open("cubed/cubed.go")
	catch(err)
	f2, err := os.Create(filepath.Join(cubedDir, "cubed.go"))
	catch(err)
	_, err = io.Copy(f2, f)
	catch(err)

	err = f2.Close()
	catch(err)
	err = f.Close()
	catch(err)
}
