package main

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
#include <stdlib.h>

#define LOG_TAG "FRP"
#define LOGD(...) __android_log_print(ANDROID_LOG_DEBUG, LOG_TAG, __VA_ARGS__)
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
	mu      sync.Mutex
	frpcCmd *exec.Cmd
	frpsCmd *exec.Cmd
)

//export Java_com_handreace_frp_FrpManager_startFrpc
func Java_com_handreace_frp_FrpManager_startFrpc(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			C.LOGD(C.CString("FRPC 已运行"))
			return
		}
		mu.Unlock()

		C.LOGD(C.CString("[FRPC] 启动中..."))
		cfg := C.GoString(path)
		cmd := exec.Command("./frpc", "-c", cfg)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			C.LOGD(C.CString("[FRPC] 启动失败: "+err.Error()))
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		C.LOGD(C.CString("[FRPC] 已启动"))

		go scanLog(stdout)
		go scanLog(stderr)

		cmd.Wait()
		C.LOGD(C.CString("[FRPC] 已停止"))

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
		C.LOGD(C.CString("[FRPC] 手动停止"))
	}
}

//export Java_com_handreace_frp_FrpManager_startFrps
func Java_com_handreace_frp_FrpManager_startFrps(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpsCmd != nil {
			mu.Unlock()
			C.LOGD(C.CString("FRPS 已运行"))
			return
		}
		mu.Unlock()

		C.LOGD(C.CString("[FRPS] 启动中..."))
		cfg := C.GoString(path)
		cmd := exec.Command("./frps", "-c", cfg)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			C.LOGD(C.CString("[FRPS] 启动失败: "+err.Error()))
			return
		}

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		C.LOGD(C.CString("[FRPS] 已启动"))
		go scanLog(stdout)
		go scanLog(stderr)

		cmd.Wait()
		C.LOGD(C.CString("[FRPS] 已停止"))

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
		C.LOGD(C.CString("[FRPS] 手动停止"))
	}
}

// 直接输出到 Android Logcat（绝对不崩溃）
func scanLog(rd io.Reader) {
	sc := bufio.NewScanner(rd)
	for sc.Scan() {
		txt := sc.Text()
		cs := C.CString(txt)
		C.LOGD(cs)
		C.free(unsafe.Pointer(cs))
	}
}

func main() {}
