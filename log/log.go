package log

import (
	"os"

	"github.com/fatih/color"
)

//  Errorf 打印错误信息,不会终止程序运行
func Errorf(format string, args ...interface{}) {
	color.Red("[ERROR] "+format, args...)
}

// Infof 一般结果输出
func Infof(format string, args ...interface{}) {
	color.Green("[INFO] "+format, args...)
}

// Debugf 打印调试信息
func Debugf(format string, args ...interface{}) {
		color.Yellow("[DEBUG] "+format, args...)

}

// Tracef 打印跟踪信息
func Tracef(format string, args ...interface{}) {
		color.Blue("[TRACE] "+format, args...)
}

// Fatalf 打印错误,直接退出
func Fatalf(format string, args ...interface{}) {
	color.Red("[Fatal] "+format, args...)
	os.Exit(1)
}
