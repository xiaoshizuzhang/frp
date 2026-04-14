package main

/*
#cgo LDFLAGS: -llog
#include <stdlib.h>
*/
import "C"

import (
	"os"
	"os/exec"
	"sync"
	"unsafe"
)

var (
	frpcCmd *exec.Cmd
	frpsCmd *exec.Cmd
	mu      sync.Mutex

	logCb    func(msg *C.char)
	statusCb func(status *C.char)
)

//export RegisterLogCallback
func RegisterLogCallback(cb uintptr) {
	logCb = *(*func(msg *C.char))(unsafe.Pointer(cb))
}

//export RegisterStatusCallback
func RegisterStatusCallback(cb uintptr) {
	statusCb = *(*func(status *C.char))(unsafe.Pointer(cb))
}

func logf(s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if logCb != nil {
		logCb(cs)
	}
}

func status(s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if statusCb != nil {
		statusCb(cs)
	}
}

//export StartFrpc
func StartFrpc(config *C.char) {
	mu.Lock()
	if frpcCmd != nil {
		mu.Unlock()
		logf("frpc already running")
		return
	}
	mu.Unlock()

	go func() {
		status("FRPC_STARTING")
		path := C.GoString(config)
		cmd := exec.Command("./frpc", "-c", path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		logf("frpc starting...")
		if err := cmd.Start(); err != nil {
			logf("frpc start failed: " + err.Error())
			status("FRPC_FAILED")
			mu.Lock()
			frpcCmd = nil
			mu.Unlock()
			return
		}

		status("FRPC_RUNNING")
		logf("frpc started")

		_ = cmd.Wait()
		status("FRPC_STOPPED")
		logf("frpc stopped")

		mu.Lock()
		frpcCmd = nil
		mu.Unlock()
	}()
}

//export StopFrpc
func StopFrpc() {
	mu.Lock()
	defer mu.Unlock()
	if frpcCmd != nil && frpcCmd.Process != nil {
		_ = frpcCmd.Process.Kill()
		frpcCmd = nil
		status("FRPC_STOPPING")
		logf("frpc stopped")
	}
}

//export StartFrps
func StartFrps(config *C.char) {
	mu.Lock()
	if frpsCmd != nil {
		mu.Unlock()
		logf("frps already running")
		return
	}
	mu.Unlock()

	go func() {
		status("FRPS_STARTING")
		path := C.GoString(config)
		cmd := exec.Command("./frps", "-c", path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		logf("frps starting...")
		if err := cmd.Start(); err != nil {
			logf("frps start failed: " + err.Error())
			status("FRPS_FAILED")
			mu.Lock()
			frpsCmd = nil
			mu.Unlock()
			return
		}

		status("FRPS_RUNNING")
		logf("frps started")

		_ = cmd.Wait()
		status("FRPS_STOPPED")
		logf("frps stopped")

		mu.Lock()
		frpsCmd = nil
		mu.Unlock()
	}()
}

//export StopFrps
func StopFrps() {
	mu.Lock()
	defer mu.Unlock()
	if frpsCmd != nil && frpsCmd.Process != nil {
		_ = frpsCmd.Process.Kill()
		frpsCmd = nil
		status("FRPS_STOPPING")
		logf("frps stopped")
	}
}

func main() {}
