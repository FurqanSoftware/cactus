package jail

import (
	"os"
	"os/exec"
	"path"
	"runtime"
	"syscall"
	"time"
)

func (j *Cell) Command(name string, args ...string) *Cmd {
	c := exec.Command(name, args...)
	c.Dir = j.Dir
	return &Cmd{
		Cmd: c,
	}
}

func (j *Cell) Create(name string) (*os.File, error) {
	name = path.Join(j.Dir, name)

	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (j *Cell) Dispose() error {
	return nil
}

func (c *Cmd) Run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := c.Cmd.Start()
	if err != nil {
		return err
	}

	ready := map[int]bool{}

	status := syscall.WaitStatus(0)
	usage := syscall.Rusage{}
	for {
		pid, err := syscall.Wait4(c.Cmd.Process.Pid, &status, 0, &usage)
		if err != nil {
			return err
		}

		if !ready[pid] {
			ready[pid] = true
		}

		c.Usages.Cpu = time.Duration(usage.Utime.Nano())
		c.Usages.Memory = uint64(usage.Maxrss)

		switch {
		case status.Exited():
			if status.ExitStatus() != 0 {
				return ExitError(status)
			}
			return nil

		case status.Signaled():
			return ExitError(status)
		}
	}

	panic("unreachable")
}
