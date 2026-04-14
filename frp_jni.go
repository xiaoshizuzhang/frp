package main

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
#include <stdlib.h>

#define LOG_TAG "FRP"
#define LOGD(...) __android_log_print(ANDROID_LOG_INFO, LOG_TAG, __VA_ARGS__)
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

// 日志输出（绝对安全，不崩溃）
func logToAndroid(msg string) {
	cs := C.CString(msg)
	C.LOGD(cs)
	C.free(unsafe.Pointer(cs))
}

// ==================== FRPC 客户端 ====================
//export Java_com_handreace_frp_FrpManager_startFrpc
func Java_com_handreace_frp_FrpManager_startFrpc(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			logToAndroid("[FRPC] 已运行")
			return
		}
		mu.Unlock()

		logToAndroid("[FRPC] 启动中...")
		cfg := C.GoString(path)
		cmd := exec.Command("./frpc", "-c", cfg)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			logToAndroid("[FRPC] 启动失败: " + err.Error())
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		logToAndroid("[FRPC] 已启动")
		go scan(stdout)
		go scan(stderr)

		cmd.Wait()
		logToAndroid("[FRPC] 已停止")

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
		logToAndroid("[FRPC] 手动停止")
	}
}

// ==================== FRPS 服务端 ====================
//export Java_com_handreace_frp_FrpManager_startFrps
func Java_com_handreace_frp_FrpManager_startFrps(env uintptr, clazz uintptr, path *C.char) {
	go func() {
		mu.Lock()
		if frpsCmd != nil {
			mu.Unlock()
			logToAndroid("[FRPS] 已运行")
			return
		}
		mu.Unlock()

		logToAndroid("[FRPS] 启动中...")
		cfg := C.GoString(path)
		cmd := exec.Command("./frps", "-c", cfg)

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			logToAndroid("[FRPS] 启动失败: " + err.Error())
			return
		}

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		logToAndroid("[FRPS] 已启动")
		go scan(stdout)
		go scan(stderr)

		cmd.Wait()
		logToAndroid("[FRPS] 已停止")

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
		logToAndroid("[FRPS] 手动停止")
	}
}

// 读取输出
func scan(r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		logToAndroid(sc.Text())
	}
}

func main() {}
