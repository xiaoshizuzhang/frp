package main

/*
#cgo LDFLAGS: -llog
#include <stdlib.h>
*/
import "C"

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"unsafe"
)

var (
	mu sync.Mutex

	frpcCmd *exec.Cmd
	frpsCmd *exec.Cmd

	// 日志和状态的函数指针
	logFunc    uintptr
	stateFunc  uintptr
)

// 注册日志回调
//export RegisterLog
func RegisterLog(cb uintptr) {
	mu.Lock()
	logFunc = cb
	mu.Unlock()
}

// 注册状态回调
//export RegisterState
func RegisterState(cb uintptr) {
	mu.Lock()
	stateFunc = cb
	mu.Unlock()
}

// 调用 Java 日志回调
func sendLog(msg string) {
	mu.Lock()
	defer mu.Unlock()
	if logFunc != 0 {
		cs := C.CString(msg)
		(*(*func(uintptr))(unsafe.Pointer(logFunc)))(uintptr(unsafe.Pointer(cs)))
		C.free(unsafe.Pointer(cs))
	}
}

// 调用 Java 状态回调
func sendState(st string) {
	mu.Lock()
	defer mu.Unlock()
	if stateFunc != 0 {
		cs := C.CString(st)
		(*(*func(uintptr))(unsafe.Pointer(stateFunc)))(uintptr(unsafe.Pointer(cs)))
		C.free(unsafe.Pointer(cs))
	}
}

//export StartFrpc
func StartFrpc(cfg *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			sendLog("frpc 正在运行")
			return
		}
		mu.Unlock()

		sendState("STARTING")
		sendLog("开始启动 frpc...")

		cmd := exec.Command("./frpc", "-c", C.GoString(cfg))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			sendLog("启动失败: " + err.Error())
			sendState("ERROR")
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		sendState("RUNNING")
		go readPipe(stdout)
		go readPipe(stderr)

		_ = cmd.Wait()
		sendState("STOPPED")
		sendLog("frpc 已停止")

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
		sendState("STOPPING")
		sendLog("手动停止 frpc")
	}
}

//export StartFrps
func StartFrps(cfg *C.char) {
	sendLog("暂未实现 frps，如需开启我可以加上")
}

//export StopFrps
func StopFrps() {}

func readPipe(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sendLog(scanner.Text())
	}
}

func main() {}
