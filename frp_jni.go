package main

/*
#cgo LDFLAGS: -llog
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/fatedier/frp/cmd/frpc"
	"github.com/fatedier/frp/cmd/frps"
)

var (
	frpcRunning bool
	frpsRunning bool
	mu          sync.Mutex

	// 日志回调函数指针（Java 层注册）
	logCallback func(msg *C.char)
	// 状态回调函数指针
	statusCallback func(status *C.char)
)

// 注册日志回调
//export RegisterLogCallback
func RegisterLogCallback(cb unsafe.Pointer) {
	logCallback = *(*func(msg *C.char))(cb)
}

// 注册状态回调
//export RegisterStatusCallback
func RegisterStatusCallback(cb unsafe.Pointer) {
	statusCallback = *(*func(status *C.char))(cb)
}

// 日志输出（自动回调 Java）
func logf(format string, args ...interface{}) {
	msg := C.CString(fmt.Sprintf(format, args...))
	if logCallback != nil {
		logCallback(msg)
	}
	C.free(unsafe.Pointer(msg))
}

// 状态回调
func sendStatus(status string) {
	cs := C.CString(status)
	if statusCallback != nil {
		statusCallback(cs)
	}
	C.free(unsafe.Pointer(cs))
}

// 启动 frpc
//export StartFrpc
func StartFrpc(configPath *C.char) {
	mu.Lock()
	if frpcRunning {
		mu.Unlock()
		logf("frpc 已经在运行")
		return
	}
	frpcRunning = true
	mu.Unlock()

	sendStatus("FRPC_STARTING")
	logf("开始启动 frpc，配置路径: %s", C.GoString(configPath))

	go func() {
		defer func() {
			mu.Lock()
			frpcRunning = false
			mu.Unlock()
			sendStatus("FRPC_STOPPED")
			logf("frpc 已停止")
		}()

		os.Args = []string{"frpc", "-c", C.GoString(configPath)}
		frpc.Main()
	}()
}

// 停止 frpc
//export StopFrpc
func StopFrpc() {
	mu.Lock()
	defer mu.Unlock()
	if !frpcRunning {
		logf("frpc 未运行")
		return
	}
	sendStatus("FRPC_STOPPING")
	logf("正在停止 frpc...")
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	frpcRunning = false
}

// 启动 frps
//export StartFrps
func StartFrps(configPath *C.char) {
	mu.Lock()
	if frpsRunning {
		mu.Unlock()
		logf("frps 已经在运行")
		return
	}
	frpsRunning = true
	mu.Unlock()

	sendStatus("FRPS_STARTING")
	logf("开始启动 frps，配置路径: %s", C.GoString(configPath))

	go func() {
		defer func() {
			mu.Lock()
			frpsRunning = false
			mu.Unlock()
			sendStatus("FRPS_STOPPED")
			logf("frps 已停止")
		}()

		os.Args = []string{"frps", "-c", C.GoString(configPath)}
		frps.Main()
	}()
}

// 停止 frps
//export StopFrps
func StopFrps() {
	mu.Lock()
	defer mu.Unlock()
	if !frpsRunning {
		logf("frps 未运行")
		return
	}
	sendStatus("FRPS_STOPPING")
	logf("正在停止 frps...")
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	frpsRunning = false
}

func main() {}
