package main

/*
#cgo LDFLAGS: -llog
#include <stdlib.h>
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/server"
)

var (
	cli    *client.Client
	svr    *server.Server
	mu     sync.Mutex
	logCb  func(msg *C.char)
	statCb func(status *C.char)
)

//export RegisterLogCallback
func RegisterLogCallback(cb uintptr) {
	logCb = *(*func(msg *C.char))(unsafe.Pointer(cb))
}

//export RegisterStatusCallback
func RegisterStatusCallback(cb uintptr) {
	statCb = *(*func(status *C.char))(unsafe.Pointer(cb))
}

func logf(format string, args ...interface{}) {
	msg := C.CString(format)
	defer C.free(unsafe.Pointer(msg))
	if logCb != nil {
		logCb(msg)
	}
}

func status(s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if statCb != nil {
		statCb(cs)
	}
}

//export StartFrpc
func StartFrpc(cfgPath *C.char) {
	mu.Lock()
	if cli != nil {
		mu.Unlock()
		logf("frpc already running")
		return
	}
	mu.Unlock()

	go func() {
		status("FRPC_STARTING")
		c, err := client.NewClientFromPath(C.GoString(cfgPath))
		if err != nil {
			logf("frpc start failed: " + err.Error())
			status("FRPC_FAILED")
			return
		}

		mu.Lock()
		cli = c
		mu.Unlock()

		logf("frpc started")
		status("FRPC_RUNNING")
		<-c.ClosedCh()
		status("FRPC_STOPPED")
		mu.Lock()
		cli = nil
		mu.Unlock()
	}()
}

//export StopFrpc
func StopFrpc() {
	mu.Lock()
	defer mu.Unlock()
	if cli != nil {
		cli.Close()
		cli = nil
		status("FRPC_STOPPING")
		logf("frpc stopped")
	}
}

//export StartFrps
func StartFrps(cfgPath *C.char) {
	mu.Lock()
	if svr != nil {
		mu.Unlock()
		logf("frps already running")
		return
	}
	mu.Unlock()

	go func() {
		status("FRPS_STARTING")
		s, err := server.NewServerFromPath(C.GoString(cfgPath))
		if err != nil {
			logf("frps start failed: " + err.Error())
			status("FRPS_FAILED")
			return
		}

		mu.Lock()
		svr = s
		mu.Unlock()

		logf("frps started")
		status("FRPS_RUNNING")
		<-s.ClosedCh()
		status("FRPS_STOPPED")
		mu.Lock()
		svr = nil
		mu.Unlock()
	}()
}

//export StopFrps
func StopFrps() {
	mu.Lock()
	defer mu.Unlock()
	if svr != nil {
		svr.Close()
		svr = nil
		status("FRPS_STOPPING")
		logf("frps stopped")
	}
}

func main() {}
