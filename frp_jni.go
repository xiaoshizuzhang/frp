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

	// 回调函数：Go -> Java
	onLog   func(line *C.char)
	onState func(state *C.char)
)

// 注册日志回调
//export SetLogCallback
func SetLogCallback(cb uintptr) {
	mu.Lock()
	onLog = *(*func(line *C.char))(unsafe.Pointer(cb))
	mu.Unlock()
}

// 注册状态回调
//export SetStateCallback
func SetStateCallback(cb uintptr) {
	mu.Lock()
	onState = *(*func(state *C.char))(unsafe.Pointer(cb))
	mu.Unlock()
}

func log(s string) {
	mu.Lock()
	defer mu.Unlock()
	if onLog != nil {
		cs := C.CString(s)
		onLog(cs)
		C.free(unsafe.Pointer(cs))
	}
}

func state(s string) {
	mu.Lock()
	defer mu.Unlock()
	if onState != nil {
		cs := C.CString(s)
		onState(cs)
		C.free(unsafe.Pointer(cs))
	}
}

//export StartFrpc
func StartFrpc(cfg *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			log("frpc 已经运行")
			return
		}
		mu.Unlock()

		state("STARTING")
		log("启动 frpc...")

		cmd := exec.Command("./frpc", "-c", C.GoString(cfg))

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			log("启动失败: " + err.Error())
			state("ERROR")
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		state("RUNNING")

		// 输出日志到 Java
		go scanPipe(stdout)
		go scanPipe(stderr)

		_ = cmd.Wait()
		state("STOPPED")
		log("frpc 已停止")

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
		state("STOPPING")
		log("正在停止 frpc...")
	}
}

//export StartFrps
func StartFrps(cfg *C.char) {
	go func() {
		mu.Lock()
		if frpsCmd != nil {
			mu.Unlock()
			log("frps 已经运行")
			return
		}
		mu.Unlock()

		state("SERVER_STARTING")
		cmd := exec.Command("./frps", "-c", C.GoString(cfg))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			log("frps 启动失败: " + err.Error())
			state("SERVER_ERROR")
			return
		}

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		state("SERVER_RUNNING")
		go scanPipe(stdout)
		go scanPipe(stderr)

		_ = cmd.Wait()
		state("SERVER_STOPPED")
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
		state("SERVER_STOPPING")
	}
}

// 读取输出并回调
func scanPipe(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log(scanner.Text())
	}
}

func main() {}
