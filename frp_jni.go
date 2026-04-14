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
)

//export StartFrpc
func StartFrpc(configPath *C.char) {
	mu.Lock()
	if frpcCmd != nil {
		mu.Unlock()
		return
	}
	mu.Unlock()

	go func() {
		path := C.GoString(configPath)
		cmd := exec.Command("./frpc", "-c", path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		_ = cmd.Run()

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
	}
}

//export StartFrps
func StartFrps(configPath *C.char) {
	mu.Lock()
	if frpsCmd != nil {
		mu.Unlock()
		return
	}
	mu.Unlock()

	go func() {
		path := C.GoString(configPath)
		cmd := exec.Command("./frps", "-c", path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		_ = cmd.Run()

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
	}
}

func main() {}
