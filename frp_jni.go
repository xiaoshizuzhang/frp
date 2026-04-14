package main

/*
#cgo LDFLAGS: -llog
#include <jni.h>
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
	mu         sync.Mutex
	frpcCmd    *exec.Cmd
	frpsCmd    *exec.Cmd

	jvm        *C.JavaVM
	logObj     C.jobject
	logMid     C.jmethodID
	stateObj   C.jobject
	stateMid   C.jmethodID
)

//export Java_com_handreace_frp_FrpManager_initCallbacks
func Java_com_handreace_frp_FrpManager_initCallbacks(
	env *C.JNIEnv,
	clazz C.jclass,
	logCb C.jobject,
	stateCb C.jobject,
) {
	// 获取 JavaVM
	(*C.JNIEnv).GetJavaVM(env, &jvm)

	// 日志回调方法
	logCls := (*C.JNIEnv).FindClass(env, C.CString("com/handreace/frp/FrpManager$LogCallback"))
	logMid = (*C.JNIEnv).GetMethodID(env, logCls, C.CString("onLog"), C.CString("(Ljava/lang/String;)V"))
	logObj = (*C.JNIEnv).NewGlobalRef(env, logCb)

	// 状态回调方法
	stateCls := (*C.JNIEnv).FindClass(env, C.CString("com/handreace/frp/FrpManager$StateCallback"))
	stateMid = (*C.JNIEnv).GetMethodID(env, stateCls, C.CString("onState"), C.CString("(Ljava/lang/String;)V"))
	stateObj = (*C.JNIEnv).NewGlobalRef(env, stateCb)
}

// 安全回调日志
func onLog(msg string) {
	if jvm == nil || logObj == nil || logMid == nil {
		return
	}

	var env *C.JNIEnv
	if C.jint(jvm.AttachCurrentThread(jvm, &env, nil)) != C.JNI_OK {
		return
	}

	cs := C.CString(msg)
	jstr := (*C.JNIEnv).NewStringUTF(env, cs)
	(*C.JNIEnv).CallVoidMethod(env, logObj, logMid, jstr)
	(*C.JNIEnv).DeleteLocalRef(env, jstr)
	C.free(unsafe.Pointer(cs))

	jvm.DetachCurrentThread(jvm)
}

// 安全回调状态
func onState(state string) {
	if jvm == nil || stateObj == nil || stateMid == nil {
		return
	}

	var env *C.JNIEnv
	if C.jint(jvm.AttachCurrentThread(jvm, &env, nil)) != C.JNI_OK {
		return
	}

	cs := C.CString(state)
	jstr := (*C.JNIEnv).NewStringUTF(env, cs)
	(*C.JNIEnv).CallVoidMethod(env, stateObj, stateMid, jstr)
	(*C.JNIEnv).DeleteLocalRef(env, jstr)
	C.free(unsafe.Pointer(cs))

	jvm.DetachCurrentThread(jvm)
}

// ==================== FRPC 客户端 ====================
//export Java_com_handreace_frp_FrpManager_startFrpc
func Java_com_handreace_frp_FrpManager_startFrpc(env *C.JNIEnv, clazz C.jclass, path *C.char) {
	go func() {
		mu.Lock()
		if frpcCmd != nil {
			mu.Unlock()
			onLog("frpc 已运行")
			return
		}
		mu.Unlock()

		onState("FRPC_STARTING")
		onLog("FRPC 启动中...")

		cmd := exec.Command("./frpc", "-c", C.GoString(path))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			onLog("FRPC 启动失败: " + err.Error())
			onState("FRPC_ERROR")
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		onState("FRPC_RUNNING")
		go readPipe(stdout)
		go readPipe(stderr)

		_ = cmd.Wait()
		onState("FRPC_STOPPED")
		onLog("FRPC 已停止")

		mu.Lock()
		frpcCmd = nil
		mu.Unlock()
	}()
}

//export Java_com_handreace_frp_FrpManager_stopFrpc
func Java_com_handreace_frp_FrpManager_stopFrpc(env *C.JNIEnv, clazz C.jclass) {
	mu.Lock()
	defer mu.Unlock()
	if frpcCmd != nil && frpcCmd.Process != nil {
		_ = frpcCmd.Process.Kill()
		frpcCmd = nil
		onState("FRPC_STOPPING")
		onLog("FRPC 已手动停止")
	}
}

// ==================== FRPS 服务端 ====================
//export Java_com_handreace_frp_FrpManager_startFrps
func Java_com_handreace_frp_FrpManager_startFrps(env *C.JNIEnv, clazz C.jclass, path *C.char) {
	go func() {
		mu.Lock()
		if frpsCmd != nil {
			mu.Unlock()
			onLog("frps 已运行")
			return
		}
		mu.Unlock()

		onState("FRPS_STARTING")
		onLog("FRPS 启动中...")

		cmd := exec.Command("./frps", "-c", C.GoString(path))
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			onLog("FRPS 启动失败: " + err.Error())
			onState("FRPS_ERROR")
			return
		}

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		onState("FRPS_RUNNING")
		go readPipe(stdout)
		go readPipe(stderr)

		_ = cmd.Wait()
		onState("FRPS_STOPPED")
		onLog("FRPS 已停止")

		mu.Lock()
		frpsCmd = nil
		mu.Unlock()
	}()
}

//export Java_com_handreace_frp_FrpManager_stopFrps
func Java_com_handreace_frp_FrpManager_stopFrps(env *C.JNIEnv, clazz C.jclass) {
	mu.Lock()
	defer mu.Unlock()
	if frpsCmd != nil && frpsCmd.Process != nil {
		_ = frpsCmd.Process.Kill()
		frpsCmd = nil
		onState("FRPS_STOPPING")
		onLog("FRPS 已手动停止")
	}
}

func readPipe(r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		onLog(sc.Text())
	}
}

func main() {}
