package main

import (
	"bufio"
	"os"
	"os/exec"
	"sync"
	"unsafe"
)

/*
#cgo LDFLAGS: -llog
*/
import "C"

var (
	mu      sync.Mutex
	frpcCmd *exec.Cmd
	frpsCmd *exec.Cmd

	// 全局保存 Java 虚拟机环境
	jvm uintptr
	// 日志回调（gobind 风格，不直接保存函数指针）
	logListenerEnv  uintptr
	logListenerObj  uintptr
	logListenerMethod uintptr
)

//export Java_com_handreace_frp_FrpManager_setLogListener
func Java_com_handreace_frp_FrpManager_setLogListener(env uintptr, clazz uintptr, obj uintptr) {
	// 这里保存全局引用（gobind 规范）
	mu.Lock()
	defer mu.Unlock()

	logListenerEnv = env
	logListenerObj = obj
	// 注意：真实生产用 gobind 生成，这里只保证不崩溃
}

func sendLog(msg string) {
	mu.Lock()
	defer mu.Unlock()

	// 不直接调用！只输出到日志，避免崩溃
	println("[FRP] " + msg)
}

// ==================== FRPC ====================
//export Java_com_handreace_frp_FrpManager_startFrpc
func Java_com_handreace_frp_FrpManager_startFrpc(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			sendLog("frpc already running")
			return
		}
		mu.Unlock()

		cmdPath := C.GoString(path)
		cmd := exec.Command(cmdPath)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			sendLog(scanner.Text())
		}

		errScanner := bufio.NewScanner(stderr)
		for errScanner.Scan() {
			sendLog(errScanner.Text())
		}

		_ = cmd.Run()

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
	}
}

// ==================== FRPS ====================
//export Java_com_handreace_frp_FrpManager_startFrps
func Java_com_handreace_frp_FrpManager_startFrps(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpsCmd != nil {
			mu.Unlock()
			sendLog("frps already running")
			return
		}
		mu.Unlock()

		cmdPath := C.GoString(path)
		cmd := exec.Command(cmdPath)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			sendLog(scanner.Text())
		}

		errScanner := bufio.NewScanner(stderr)
		for errScanner.Scan() {
			sendLog(errScanner.Text())
		}

		_ = cmd.Run()

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
	}
}

func main() {
	// 空
}
