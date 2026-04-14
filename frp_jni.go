package main

/*
#cgo LDFLAGS: -llog
#include <jni.h>
#include <stdlib.h>
#include <android/log.h>

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
	mu         sync.Mutex
	frpcCmd    *exec.Cmd
	frpsCmd    *exec.Cmd

	jvm        *C.JavaVM
	logObj     C.jobject  // 全局引用
	logMethod  C.jmethodID
	stateObj   C.jobject
	stateMethod C.jmethodID
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

	// 日志回调
	logCls := (*C.JNIEnv).FindClass(env, C.CString("com/handreace/frp/FrpManager$LogCallback"))
	logMethod = (*C.JNIEnv).GetMethodID(env, logCls, C.CString("onLog"), C.CString("(Ljava/lang/String;)V"))
	logObj = (*C.JNIEnv).NewGlobalRef(env, logCb)

	// 状态回调
	stateCls := (*C.JNIEnv).FindClass(env, C.CString("com/handreace/frp/FrpManager$StateCallback"))
	stateMethod = (*C.JNIEnv).GetMethodID(env, stateCls, C.CString("onState"), C.CString("(Ljava/lang/String;)V"))
	stateObj = (*C.JNIEnv).NewGlobalRef(env, stateCb)

	(*C.JNIEnv).DeleteLocalRef(env, logCls)
	(*C.JNIEnv).DeleteLocalRef(env, stateCls)
}

// 安全回调 Java
func onLog(msg string) {
	if jvm == nil || logObj == nil || logMethod == nil {
		C.LOGD(C.CString(msg))
		return
	}

	var env *C.JNIEnv
	if C.jint(jvm.AttachCurrentThread(jvm, &env, nil)) != C.JNI_OK {
		return
	}
	cstr := C.CString(msg)
	jstr := (*C.JNIEnv).NewStringUTF(env, cstr)
	(*C.JNIEnv).CallVoidMethod(env, logObj, logMethod, jstr)
	(*C.JNIEnv).DeleteLocalRef(env, jstr)
	C.free(unsafe.Pointer(cstr))
	jvm.DetachCurrentThread(jvm)
}

func onState(state string) {
	if jvm == nil || stateObj == nil || stateMethod == nil {
		return
	}

	var env *C.JNIEnv
	if C.jint(jvm.AttachCurrentThread(jvm, &env, nil)) != C.JNI_OK {
		return
	}
	cstr := C.CString(state)
	jstr := (*C.JNIEnv).NewStringUTF(env, cstr)
	(*C.JNIEnv).CallVoidMethod(env, stateObj, stateMethod, jstr)
	(*C.JNIEnv).DeleteLocalRef(env, jstr)
	C.free(unsafe.Pointer(cstr))
	jvm.DetachCurrentThread(jvm)
}

// ==================== FRPC ====================
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
		onLog("启动 frpc...")

		cfg := C.GoString(path)
		cmd := exec.Command("./frpc", "-c", cfg)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			onLog("启动失败: " + err.Error())
			onState("FRPC_ERROR")
			return
		}

		mu.Lock()
		frpcCmd = cmd
		mu.Unlock()

		onState("FRPC_RUNNING")
		go scan(stdout)
		go scan(stderr)

		cmd.Wait()
		onState("FRPC_STOPPED")
		onLog("frpc 已停止")

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
		onLog("已停止 frpc")
	}
}

// ==================== FRPS ====================
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
		onLog("启动 frps...")

		cfg := C.GoString(path)
		cmd := exec.Command("./frps", "-c", cfg)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			onLog("启动失败: " + err.Error())
			onState("FRPS_ERROR")
			return
		}

		mu.Lock()
		frpsCmd = cmd
		mu.Unlock()

		onState("FRPS_RUNNING")
		go scan(stdout)
		go scan(stderr)

		cmd.Wait()
		onState("FRPS_STOPPED")
		onLog("frps 已停止")

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
		onLog("已停止 frps")
	}
}

func scan(r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		onLog(sc.Text())
	}
}

func main() {}
