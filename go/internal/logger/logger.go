package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Logger はログ出力を管理する構造体
type Logger struct {
	enabled bool
	logger  *log.Logger
}

var defaultLogger = &Logger{
	enabled: false,
	logger:  log.New(os.Stderr, "", 0),
}

// Enable はログ出力を有効化する
func Enable() {
	defaultLogger.enabled = true
}

// Disable はログ出力を無効化する
func Disable() {
	defaultLogger.enabled = false
}

// IsEnabled はログ出力が有効かどうかを返す
func IsEnabled() bool {
	return defaultLogger.enabled
}

// SetOutput はログの出力先を設定する
func SetOutput(w io.Writer) {
	defaultLogger.logger.SetOutput(w)
}

// Info は情報ログを出力する
func Info(format string, v ...interface{}) {
	if defaultLogger.enabled {
		defaultLogger.logger.Printf("[INFO] "+format, v...)
	}
}

// Debug はデバッグログを出力する
func Debug(format string, v ...interface{}) {
	if defaultLogger.enabled {
		defaultLogger.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Warn は警告ログを出力する
func Warn(format string, v ...interface{}) {
	if defaultLogger.enabled {
		defaultLogger.logger.Printf("[WARN] "+format, v...)
	}
}

// Error はエラーログを出力する（enabledに関わらず常に出力）
func Error(format string, v ...interface{}) {
	defaultLogger.logger.Printf("[ERROR] "+format, v...)
}

// Infof は改行なしの情報ログを出力する
func Infof(format string, v ...interface{}) {
	if defaultLogger.enabled {
		fmt.Fprintf(defaultLogger.logger.Writer(), "[INFO] "+format, v...)
	}
}

// Section はセクション区切りを出力する
func Section(title string) {
	if defaultLogger.enabled {
		defaultLogger.logger.Printf("=== %s ===", title)
	}
}
