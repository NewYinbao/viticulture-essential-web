package sharing

import (
	"os/exec"
	"syscall"
)

func prepare(c *exec.Cmd)                 { c.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL} }
func contain(c *exec.Cmd) (func(), error) { return func() {}, nil }
