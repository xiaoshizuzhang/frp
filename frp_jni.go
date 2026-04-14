package main

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"unsafe"
)

import "C"

var (
	mu      sync.Mutex
	frpcCmd *exec.Cmd
	frpsCmd *exec.Cmd

	// gobind 官方回调方式
	onLog   func(string)
	onState func(string)
)

//export Java_com_handreace_frp_FrpManager_setLogCallback
func Java_com_handreace_frp_FrpManager_setLogCallback(env uintptr, clazz uintptr, cb uintptr) {
	onLog = *(*func(string))(unsafe.Pointer(cb))
}

//export Java_com_handreace_frp_FrpManager_setStateCallback
func Java_com_handreace_frp_FrpManager_setStateCallback(env uintptr, clazz uintptr, cb uintptr) {
	onState = *(*func(string))(unsafe.Pointer(cb))
}

func sendLog(msg string) {
	defer func() { recover() }()
	if onLog != nil {
		onLog(msg)
	}
}
func sendState(st string) {
	defer func() { recover() }()
	if onState != nil {
		onState(st)
	}
}

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
		sendLog("启动中...")

		cmd := exec.Command("./frpc", "-c", C.GoString(path))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			sendLog("启动失败: " + err.Error())
			sendState("FRPC_ERROR")
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		sendState("FRPC_RUNNING")
		scan(stdout)
		scan(stderr)

		cmd.Wait()
		sendState("FRPC_STOPPED")
		sendLog("已停止")

		mu.Lock()
		frpcCmd = nil
		mu.Unlock()
	}()
}

//export Java_com_handreace_frp_FrpManager_stopFrpc
func Java_com_handreace_frp_FrpManager_stopFrpc(env uintptr, clazz uintptr) {
	mu.Lock()
	defer mu.Unlock()
	if frpcCmd != nil {
		_ = frpcCmd.Process.Kill()
		frpcCmd = nil
		sendState("FRPC_STOPPING")
		sendLog("手动停止")
	}
}

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
		sendLog("启动中...")

		cmd := exec.Command("./frps", "-c", C.GoString(path))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			sendLog("启动失败: " + err.Error())
			sendState("FRPS_ERROR")
			return
		}

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		sendState("FRPS_RUNNING")
		scan(stdout)
		scan(stderr)

		cmd.Wait()
		sendState("FRPS_STOPPED")
		sendLog("已停止")

		mu.Lock()
		frpsCmd = nil
		mu.Unlock()
	}()
}

//export Java_com_handreace_frp_FrpManager_stopFrps
func Java_com_handreace_frp_FrpManager_stopFrps(env uintptr, clazz uintptr) {
	mu.Lock()
	defer mu.Unlock()
	if frpsCmd != nil {
		_ = frpsCmd.Process.Kill()
		frpsCmd = nil
		sendState("FRPS_STOPPING")
		sendLog("手动停止")
	}
}

func scan(r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		sendLog(sc.Text())
	}
}

func main() {}
