package jail

import (
	"os"
	"os/exec"
	"path"
	"runtime"
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

	state, err := c.Cmd.Process.Wait()
	if err != nil {
		return err
	}
	c.Usages.Cpu = state.UserTime()

	return nil
}
