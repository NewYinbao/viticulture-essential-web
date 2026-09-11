package sharing

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

var kernel = syscall.NewLazyDLL("kernel32.dll")

func prepare(c *exec.Cmd) {}

// Kernel-owned job kills only this child when parent handles close, including console close.
func contain(c *exec.Cmd) (func(), error) {
	job, _, e := kernel.NewProc("CreateJobObjectW").Call(0, 0)
	if job == 0 {
		return nil, e
	}
	closeJob := func() { syscall.CloseHandle(syscall.Handle(job)) }
	type basic struct {
		ProcessTime, JobTime int64
		Flags                uint32
		Min, Max             uintptr
		Active               uint32
		Affinity             uintptr
		Priority, Scheduling uint32
	}
	type info struct {
		Basic                                          basic
		IO                                             [6]uint64
		ProcessMemory, JobMemory, PeakProcess, PeakJob uintptr
	}
	limits := info{}
	limits.Basic.Flags = 0x2000
	ok, _, e := kernel.NewProc("SetInformationJobObject").Call(job, 9, uintptr(unsafe.Pointer(&limits)), unsafe.Sizeof(limits))
	if ok == 0 {
		closeJob()
		return nil, fmt.Errorf("job limits: %w", e)
	}
	process, e := syscall.OpenProcess(0x0100|0x0001, false, uint32(c.Process.Pid))
	if e != nil {
		closeJob()
		return nil, e
	}
	defer syscall.CloseHandle(process)
	ok, _, e = kernel.NewProc("AssignProcessToJobObject").Call(job, uintptr(process))
	if ok == 0 {
		closeJob()
		return nil, fmt.Errorf("job assignment: %w", e)
	}
	return closeJob, nil
}
