// +build !windows

package baker

import (
	"syscall"
)

func bakerSyscall(trap, nargs, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall(trap, a1, a2, a3)
}
