package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAcquireLock(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test.state")

	// ロック取得
	lock, err := AcquireLock(stateFile, 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	defer lock.Release()

	if lock == nil {
		t.Fatal("AcquireLock() returned nil")
	}

	// ロックファイルが存在することを確認
	lockFile := stateFile + ".lock"
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file should exist")
	}
}

func TestAcquireLock_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test.state")

	// 1つ目のロックを取得
	lock1, err := AcquireLock(stateFile, 5*time.Second)
	if err != nil {
		t.Fatalf("First AcquireLock() error = %v", err)
	}
	defer lock1.Release()

	// 2つ目のロックを取得を試みる（タイムアウトするはず）
	start := time.Now()
	lock2, err := AcquireLock(stateFile, 1*time.Second)
	elapsed := time.Since(start)

	if err == nil {
		lock2.Release()
		t.Fatal("Second AcquireLock() should fail with timeout")
	}

	// タイムアウト時間が正しいことを確認（誤差を考慮）
	if elapsed < 900*time.Millisecond || elapsed > 1500*time.Millisecond {
		t.Errorf("Timeout elapsed = %v, want ~1s", elapsed)
	}
}

func TestFileLock_Release(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test.state")
	lockFile := stateFile + ".lock"

	// ロック取得
	lock, err := AcquireLock(stateFile, 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}

	// ロック解放
	if err := lock.Release(); err != nil {
		t.Errorf("Release() error = %v", err)
	}

	// ロックファイルが削除されていることを確認
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("Lock file should be removed after release")
	}
}

func TestFileLock_Release_AlreadyReleased(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test.state")

	// ロック取得
	lock, err := AcquireLock(stateFile, 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}

	// 1回目のリリース
	if err := lock.Release(); err != nil {
		t.Errorf("First Release() error = %v", err)
	}

	// 2回目のリリース（エラーにならないこと）
	if err := lock.Release(); err != nil {
		t.Errorf("Second Release() should not error, got %v", err)
	}
}

func TestAcquireLock_Sequential(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test.state")

	// 1つ目のロックを取得して解放
	lock1, err := AcquireLock(stateFile, 5*time.Second)
	if err != nil {
		t.Fatalf("First AcquireLock() error = %v", err)
	}
	if err := lock1.Release(); err != nil {
		t.Fatalf("First Release() error = %v", err)
	}

	// 2つ目のロックを取得（成功するはず）
	lock2, err := AcquireLock(stateFile, 5*time.Second)
	if err != nil {
		t.Fatalf("Second AcquireLock() error = %v", err)
	}
	defer lock2.Release()

	// 成功
}

func TestAcquireLock_WritesProcessID(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test.state")
	lockFile := stateFile + ".lock"

	// ロック取得
	lock, err := AcquireLock(stateFile, 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	defer lock.Release()

	// ロックファイルにプロセスIDが書き込まれていることを確認
	content, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if len(content) == 0 {
		t.Error("Lock file should contain process ID")
	}
}
