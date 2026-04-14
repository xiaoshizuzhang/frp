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

	logCB    func(string)
	statusCB func(string)
)

//export Java_com_handreace_frp_FrpManager_registerLogCallback
func Java_com_handreace_frp_FrpManager_registerLogCallback(env uintptr, clazz uintptr, cb uintptr) {
	logCB = func(s string) {
		cs := C.CString(s)
		CallJavaVoidMethod(env, cb, cs)
		C.free(unsafe.Pointer(cs))
	}
}

//export Java_com_handreace_frp_FrpManager_registerStatusCallback
func Java_com_handreace_frp_FrpManager_registerStatusCallback(env uintptr, clazz uintptr, cb uintptr) {
	statusCB = func(s string) {
		cs := C.CString(s)
		CallJavaVoidMethod(env, cb, cs)
		C.free(unsafe.Pointer(cs))
	}
}

//export Java_com_handreace_frp_FrpManager_startFrpc
func Java_com_handreace_frp_FrpManager_startFrpc(env uintptr, clazz uintptr, path *C.char) {
	go startFrpc(C.GoString(path))
}

//export Java_com_handreace_frp_FrpManager_stopFrpc
func Java_com_handreace_frp_FrpManager_stopFrpc(env uintptr, clazz uintptr) {
	stopFrpc()
}

//export Java_com_handreace_frp_FrpManager_startFrps
func Java_com_handreace_frp_FrpManager_startFrps(env uintptr, clazz uintptr, path *C.char) {
	go startFrps(C.GoString(path))
}

//export Java_com_handreace_frp_FrpManager_stopFrps
func Java_com_handreace_frp_FrpManager_stopFrps(env uintptr, clazz uintptr) {
	stopFrps()
}

func logf(s string) {
	if logCB != nil {
		logCB(s)
	}
}

func sendStatus(s string) {
	if statusCB != nil {
		statusCB(s)
	}
}

func startFrpc(path string) {
	mu.Lock()
	if frpcCmd != nil {
		mu.Unlock()
		logf("frpc already running")
		return
	}
	mu.Unlock()

	sendStatus("FRPC_STARTING")
	cmd := exec.Command("/system/bin/sh", "-c", "frpc -c "+path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	mu.Lock()
	frpcCmd = cmd
	mu.Unlock()

	if err := cmd.Start(); err != nil {
		logf("frpc start failed: " + err.Error())
		sendStatus("FRPC_FAILED")
		mu.Lock()
		frpcCmd = nil
		mu.Unlock()
		return
	}

	sendStatus("FRPC_RUNNING")
	logf("frpc started")
	_ = cmd.Wait()
	sendStatus("FRPC_STOPPED")

	mu.Lock()
	frpcCmd = nil
	mu.Unlock()
}

func stopFrpc() {
	mu.Lock()
	defer mu.Unlock()
	if frpcCmd != nil && frpcCmd.Process != nil {
		_ = frpcCmd.Process.Kill()
		frpcCmd = nil
		sendStatus("FRPC_STOPPING")
		logf("frpc stopped")
	}
}

func startFrps(path string) {
	mu.Lock()
	if frpsCmd != nil {
		mu.Unlock()
		logf("frps already running")
		return
	}
	mu.Unlock()

	sendStatus("FRPS_STARTING")
	cmd := exec.Command("/system/bin/sh", "-c", "frps -c "+path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	mu.Lock()
	frpsCmd = cmd
	mu.Unlock()

	if err := cmd.Start(); err != nil {
		logf("frps start failed: " + err.Error())
		sendStatus("FRPS_FAILED")
		mu.Lock()
		frpsCmd = nil
		mu.Unlock()
		return
	}

	sendStatus("FRPS_RUNNING")
	logf("frps started")
	_ = cmd.Wait()
	sendStatus("FRPS_STOPPED")

	mu.Lock()
	frpsCmd = nil
	mu.Unlock()
}

func stopFrps() {
	mu.Lock()
	defer mu.Unlock()
	if frpsCmd != nil && frpsCmd.Process != nil {
		_ = frpsCmd.Process.Kill()
		frpsCmd = nil
		sendStatus("FRPS_STOPPING")
		logf("frps stopped")
	}
}

// 空实现，避免报错
func CallJavaVoidMethod(env uintptr, method uintptr, arg *C.char) {}

func main() {}
