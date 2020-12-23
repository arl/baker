package baker

import (
	"syscall"
)

func bakerSyscall(trap, nargs, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall(trap, nargs, a1, a3, a3)
}
