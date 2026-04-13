package main

/*
# 告诉 Go 编译成 Android 可加载的 JNI 库
#cgo LDFLAGS: -llog
*/
import "C"

import (
	"os"
	"strings"

	"github.com/fatedier/frp/cmd/frpc"
	"github.com/fatedier/frp/cmd/frps"
)

// 导出 JNI 可调用函数：启动 frpc
//export StartFrpc
func StartFrpc(configPath *C.char) {
	os.Args = []string{"frpc", "-c", C.GoString(configPath)}
	go frpc.Main()
}

// 导出 JNI 可调用函数：启动 frps
//export StartFrps
func StartFrps(configPath *C.char) {
	os.Args = []string{"frps", "-c", C.GoString(configPath)}
	go frps.Main()
}

// 必须留空 main 函数，Go 编译共享库要求
func main() {}
