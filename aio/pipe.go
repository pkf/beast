// +build !windows

package aio

import (
	"os"
	"syscall"
)

// Pipe returns a connected pair of Files; reads from
// r return bytes written to w. It returns the files and an error, if any.
// Optionally, r or w might be set to non-blocking mode using the appropriate
// flags. To obtain a blocking pipe just pass 0 as the flag.
func Pipe() (r int, w int, err error) {
	var p [2]int

	syscall.ForkLock.RLock()
	if err := syscall.Pipe(p[:]); err != nil {
		syscall.ForkLock.RUnlock()
		return 0, 0, os.NewSyscallError("pipe", err)
	}
	syscall.CloseOnExec(p[0])
	syscall.CloseOnExec(p[1])

	syscall.SetNonblock(p[0], true)

	syscall.SetNonblock(p[1], true)

	syscall.ForkLock.RUnlock()

	return p[0], p[1], nil
	//return os.NewFile(uintptr(p[0]), "|0"), os.NewFile(uintptr(p[1]), "|1"), nil
}
