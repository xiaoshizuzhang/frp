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
	mu          sync.Mutex
	frpcCmd     *exec.Cmd
	frpsCmd     *exec.Cmd

	// Java 回调（官方支持方式）
	onLogFunc   func(string)
	onStateFunc func(string)
)

// ==================== 回调注册 ====================
//export Java_com_handreace_frp_FrpManager_setLogCallback
func Java_com_handreace_frp_FrpManager_setLogCallback(env uintptr, clazz uintptr, cb uintptr) {
	mu.Lock()
	onLogFunc = *(*func(string))(unsafe.Pointer(cb))
	mu.Unlock()
}

//export Java_com_handreace_frp_FrpManager_setStateCallback
func Java_com_handreace_frp_FrpManager_setStateCallback(env uintptr, clazz uintptr, cb uintptr) {
	mu.Lock()
	onStateFunc = *(*func(string))(unsafe.Pointer(cb))
	mu.Unlock()
}

// 发送日志到界面
func sendLog(s string) {
	mu.Lock()
	defer mu.Unlock()
	if onLogFunc != nil {
		onLogFunc(s)
	}
}

// 发送状态到界面
func sendState(s string) {
	mu.Lock()
	defer mu.Unlock()
	if onStateFunc != nil {
		onStateFunc(s)
	}
}

// ==================== FRPC 客户端 ====================
//export Java_com_handreace_frp_FrpManager_startFrpc
func Java_com_handreace_frp_FrpManager_startFrpc(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			sendLog("frpc 已运行")
			return
		}
		mu.Unlock()

		sendState("FRPC_STARTING")
		sendLog("FRPC 启动中...")

		cmd := exec.Command("./frpc", "-c", C.GoString(path))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			sendLog("FRPC 启动失败: " + err.Error())
			sendState("FRPC_ERROR")
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		sendState("FRPC_RUNNING")
		go scanLogs(stdout)
		go scanLogs(stderr)

		cmd.Wait()
		sendState("FRPC_STOPPED")
		sendLog("FRPC 已停止")

		mu.Lock()
		frpcCmd = nil
		mu.Unlock()
	}()
}

//export Java_com_handreace_frp_FrpManager_stopFrpc
func Java_com_handreace_frp_FrpManager_stopFrpc(env uintptr, clazz uintptr) {
	mu.Lock()
	defer mu.Unlock()
	if frpcCmd != nil && frpcCmd.Process != nil {
		_ = frpcCmd.Process.Kill()
		frpcCmd = nil
		sendState("FRPC_STOPPING")
		sendLog("FRPC 手动停止")
	}
}

// ==================== FRPS 服务端 ====================
//export Java_com_handreace_frp_FrpManager_startFrps
func Java_com_handreace_frp_FrpManager_startFrps(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpsCmd != nil {
			mu.Unlock()
			sendLog("frps 已运行")
			return
		}
		mu.Unlock()

		sendState("FRPS_STARTING")
		sendLog("FRPS 启动中...")

		cmd := exec.Command("./frps", "-c", C.GoString(path))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			sendLog("FRPS 启动失败: " + err.Error())
			sendState("FRPS_ERROR")
			return
		}

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		sendState("FRPS_RUNNING")
		go scanLogs(stdout)
		go scanLogs(stderr)

		cmd.Wait()
		sendState("FRPS_STOPPED")
		sendLog("FRPS 已停止")

		mu.Lock()
		frpsCmd = nil
		mu.Unlock()
	}()
}

//export Java_com_handreace_frp_FrpManager_stopFrps
func Java_com_handreace_frp_FrpManager_stopFrps(env uintptr, clazz uintptr) {
	mu.Lock()
	defer mu.Unlock()
	if frpsCmd != nil && frpsCmd.Process != nil {
		_ = frpsCmd.Process.Kill()
		frpsCmd = nil
		sendState("FRPS_STOPPING")
		sendLog("FRPS 手动停止")
	}
}

// 读取输出
func scanLogs(rd io.Reader) {
	sc := bufio.NewScanner(rd)
	for sc.Scan() {
		sendLog(sc.Text())
	}
}

func main() {}
