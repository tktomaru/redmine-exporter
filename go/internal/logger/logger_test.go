package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestEnableDisable(t *testing.T) {
	// 初期状態は無効
	if IsEnabled() {
		t.Error("Expected logger to be disabled by default")
	}

	// 有効化
	Enable()
	if !IsEnabled() {
		t.Error("Expected logger to be enabled after Enable()")
	}

	// 無効化
	Disable()
	if IsEnabled() {
		t.Error("Expected logger to be disabled after Disable()")
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	// 無効時は出力されない
	Disable()
	Info("test message")
	if buf.Len() != 0 {
		t.Errorf("Expected no output when disabled, got: %s", buf.String())
	}

	// 有効時は出力される
	buf.Reset()
	Enable()
	Info("test message: %s", "hello")
	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected [INFO] prefix, got: %s", output)
	}
	if !strings.Contains(output, "test message: hello") {
		t.Errorf("Expected message content, got: %s", output)
	}

	// クリーンアップ
	Disable()
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	Enable()
	Debug("debug message: %d", 42)
	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Expected [DEBUG] prefix, got: %s", output)
	}
	if !strings.Contains(output, "debug message: 42") {
		t.Errorf("Expected message content, got: %s", output)
	}

	Disable()
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	Enable()
	Warn("warning message")
	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("Expected [WARN] prefix, got: %s", output)
	}
	if !strings.Contains(output, "warning message") {
		t.Errorf("Expected message content, got: %s", output)
	}

	Disable()
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	// Errorは無効時でも出力される
	Disable()
	Error("error message")
	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Expected [ERROR] prefix, got: %s", output)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("Expected message content, got: %s", output)
	}
}

func TestSection(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil)

	Enable()
	Section("Test Section")
	output := buf.String()
	if !strings.Contains(output, "=== Test Section ===") {
		t.Errorf("Expected section format, got: %s", output)
	}

	Disable()
}
