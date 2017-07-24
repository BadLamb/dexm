package dapp

/*
#cgo LDFLAGS: -lseccomp
#include "sandbox.h"
*/
import "C"

import (
	"errors"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

var ipcPages map[int]unsafe.Pointer

func StartDApp(file string) (int, error) {
	if ipcPages == nil {
		ipcPages = make(map[int]unsafe.Pointer)
	}

	var pid C.int

	shared := C.start_app(C.CString(file), 1024, &pid)

	if uintptr(shared) == 0 {
		return 0, errors.New("mmap() failed in start_app")
	}

	log.Debug("Allocated shared memory page at ", shared)

	ipcPages[int(pid)] = shared

	return int(pid), nil
}
