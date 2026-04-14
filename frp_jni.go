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
	mu              sync.Mutex
	frpcCmd         *exec.Cmd
	frpsCmd         *exec.Cmd
	javaLogFunc     uintptr
	javaStateFunc   uintptr
)

//export Java_com_handreace_frp_FrpManager_setLogCallback
func Java_com_handreace_frp_FrpManager_setLogCallback(env uintptr, clazz uintptr, cb uintptr) {
	mu.Lock()
	javaLogFunc = cb
	mu.Unlock()
}

//export Java_com_handreace_frp_FrpManager_setStateCallback
func Java_com_handreace_frp_FrpManager_setStateCallback(env uintptr, clazz uintptr, cb uintptr) {
	mu.Lock()
	javaStateFunc = cb
	mu.Unlock()
}

func sendLog(msg string) {
	mu.Lock()
	defer mu.Unlock()
	if javaLogFunc == 0 {
		return
	}
	cs := C.CString(msg)
	(*(*func(uintptr))(unsafe.Pointer(javaLogFunc)))(uintptr(unsafe.Pointer(cs)))
	C.free(unsafe.Pointer(cs))
}

func sendState(state string) {
	mu.Lock()
	defer mu.Unlock()
	if javaStateFunc == 0 {
		return
	}
	cs := C.CString(state)
	(*(*func(uintptr))(unsafe.Pointer(javaStateFunc)))(uintptr(unsafe.Pointer(cs)))
	C.free(unsafe.Pointer(cs))
}

//export Java_com_handreace_frp_FrpManager_startFrpc
func Java_com_handreace_frp_FrpManager_startFrpc(env uintptr, clazz uintptr, cfgPath *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			sendLog("FRPC 已经运行")
			return
		}
		mu.Unlock()

		sendState("FRPC_STARTING")
		sendLog("FRPC 启动中...")

		cmd := exec.Command("./frpc", "-c", C.GoString(cfgPath))
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
		go readPipe(stdout)
		go readPipe(stderr)

		_ = cmd.Wait()
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

//export Java_com_handreace_frp_FrpManager_startFrps
func Java_com_handreace_frp_FrpManager_startFrps(env uintptr, clazz uintptr, cfgPath *C.char) {
	go func() {
		mu.Lock()
		if frpsCmd != nil {
			mu.Unlock()
			sendLog("FRPS 已经运行")
			return
		}
		mu.Unlock()

		sendState("FRPS_STARTING")
		sendLog("FRPS 启动中...")

		cmd := exec.Command("./frps", "-c", C.GoString(cfgPath))
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
		go readPipe(stdout)
		go readPipe(stderr)

		_ = cmd.Wait()
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

func readPipe(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sendLog(scanner.Text())
	}
}

func main() {}
