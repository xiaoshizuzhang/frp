package main

/*
#cgo LDFLAGS: -llog
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/fatedier/frp/cmd/frpc"
	"github.com/fatedier/frp/cmd/frps"
)

var (
	frpcRunning bool
	frpsRunning bool
	mu          sync.Mutex

	logCallback    func(msg *C.char)
	statusCallback func(status *C.char)
)

//export RegisterLogCallback
func RegisterLogCallback(cb uintptr) {
	logCallback = *(*func(msg *C.char))(unsafe.Pointer(cb))
}

//export RegisterStatusCallback
func RegisterStatusCallback(cb uintptr) {
	statusCallback = *(*func(status *C.char))(unsafe.Pointer(cb))
}

func logf(format string, args ...interface{}) {
	msg := C.CString(fmt.Sprintf(format, args...))
	defer C.free(unsafe.Pointer(msg))
	if logCallback != nil {
		logCallback(msg)
	}
}

func sendStatus(status string) {
	cs := C.CString(status)
	defer C.free(unsafe.Pointer(cs))
	if statusCallback != nil {
		statusCallback(cs)
	}
}

//export StartFrpc
func StartFrpc(configPath *C.char) {
	mu.Lock()
	if frpcRunning {
		mu.Unlock()
		logf("frpc already running")
		return
	}
	frpcRunning = true
	mu.Unlock()

	sendStatus("FRPC_STARTING")
	go func() {
		defer func() {
			mu.Lock()
			frpcRunning = false
			mu.Unlock()
			sendStatus("FRPC_STOPPED")
		}()
		os.Args = []string{"frpc", "-c", C.GoString(configPath)}
		frpc.Main()
	}()
}

//export StopFrpc
func StopFrpc() {
	mu.Lock()
	defer mu.Unlock()
	if !frpcRunning {
		return
	}
	sendStatus("FRPC_STOPPING")
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}

//export StartFrps
func StartFrps(configPath *C.char) {
	mu.Lock()
	if frpsRunning {
		mu.Unlock()
		logf("frps already running")
		return
	}
	frpsRunning = true
	mu.Unlock()

	sendStatus("FRPS_STARTING")
	go func() {
		defer func() {
			mu.Lock()
			frpsRunning = false
			mu.Unlock()
			sendStatus("FRPS_STOPPED")
		}()
		os.Args = []string{"frps", "-c", C.GoString(configPath)}
		frps.Main()
	}()
}

//export StopFrps
func StopFrps() {
	mu.Lock()
	defer mu.Unlock()
	if !frpsRunning {
		return
	}
	sendStatus("FRPS_STOPPING")
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}

func main() {}
