package main

/*
#cgo LDFLAGS: -llog
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/server"
)

var (
	cli        *client.Client
	svr        *server.Server
	mu         sync.Mutex
	logCb      func(msg *C.char)
	statusCb   func(status *C.char)
)

//export RegisterLogCallback
func RegisterLogCallback(cb uintptr) {
	logCb = *(*func(msg *C.char))(unsafe.Pointer(cb))
}

//export RegisterStatusCallback
func RegisterStatusCallback(cb uintptr) {
	statusCb = *(*func(status *C.char))(unsafe.Pointer(cb))
}

func logf(format string, args ...interface{}) {
	msg := C.CString(fmt.Sprintf(format, args...))
	defer C.free(unsafe.Pointer(msg))
	if logCb != nil {
		logCb(msg)
	}
}

func status(s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	if statusCb != nil {
		statusCb(cs)
	}
}

// 从 ini 路径启动 frpc
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
		path := C.GoString(cfgPath)

		cfg, pxyCfgs, _, err := config.LoadClientConfig(path)
		if err != nil {
			logf("config error: %v", err)
			status("FRPC_FAILED")
			return
		}

		c, err := client.NewClient(cfg, pxyCfgs, nil)
		if err != nil {
			logf("client error: %v", err)
			status("FRPC_FAILED")
			return
		}

		mu.Lock()
		cli = c
		mu.Unlock()

		logf("frpc started")
		status("FRPC_RUNNING")

		<-c.ClosedCh()
		logf("frpc closed")
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
		logf("frpc stop success")
		status("FRPC_STOPPING")
	}
}

// 从 ini 路径启动 frps
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
		path := C.GoString(cfgPath)

		cfg, err := config.LoadServerConfig(path)
		if err != nil {
			logf("server config err: %v", err)
			status("FRPS_FAILED")
			return
		}

		s, err := server.NewServer(cfg)
		if err != nil {
			logf("server err: %v", err)
			status("FRPS_FAILED")
			return
		}

		mu.Lock()
		svr = s
		mu.Unlock()

		logf("frps started")
		status("FRPS_RUNNING")

		<-s.ClosedCh()
		logf("frps closed")
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
		logf("frps stop success")
		status("FRPS_STOPPING")
	}
}

func main() {}
