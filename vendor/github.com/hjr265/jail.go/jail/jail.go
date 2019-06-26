package jail

import (
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

type Cell struct {
	Dir string
}

type Cmd struct {
	*exec.Cmd

	Limits Resources
	Usages Resources
}

type Resources struct {
	Cpu    time.Duration
	Memory uint64
}

type ExitError syscall.WaitStatus

func (e ExitError) Error() string {
	w := syscall.WaitStatus(e)
	s := ""

	switch {
	case w.Exited():
		s = "exit status " + strconv.Itoa(w.ExitStatus())

	case w.Signaled():
		s = "signal: " + w.Signal().String()

	case w.Stopped():
		s = "stop signal: " + w.StopSignal().String()
		if w.StopSignal() == syscall.SIGTRAP && w.TrapCause() != 0 {
			s += " (trap " + strconv.Itoa(w.TrapCause()) + ")"
		}

	case w.Continued():
		s = "continued"
	}

	if w.CoreDump() {
		s += " (core dumped)"
	}

	return s
}
