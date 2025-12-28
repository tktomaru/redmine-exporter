package state

import (
	"fmt"
	"os"
	"time"
)

// FileLock はファイルロックを管理
type FileLock struct {
	file     *os.File
	filePath string
}

// AcquireLock はファイルロックを取得
// タイムアウト付きでロック取得を試みる
func AcquireLock(filePath string, timeout time.Duration) (*FileLock, error) {
	lockFile := filePath + ".lock"

	start := time.Now()
	for {
		// ロックファイルを排他的に作成を試みる
		file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			// ロック取得成功
			// プロセスIDを書き込む（デバッグ用）
			fmt.Fprintf(file, "%d\n", os.Getpid())
			file.Sync()

			return &FileLock{
				file:     file,
				filePath: lockFile,
			}, nil
		}

		// タイムアウトチェック
		if time.Since(start) >= timeout {
			return nil, fmt.Errorf("ロック取得タイムアウト（%v経過）: 他のプロセスが実行中の可能性があります", timeout)
		}

		// 少し待ってリトライ
		time.Sleep(100 * time.Millisecond)
	}
}

// Release はファイルロックを解放
func (fl *FileLock) Release() error {
	if fl.file == nil {
		return nil
	}

	// ファイルを閉じる
	if err := fl.file.Close(); err != nil {
		return fmt.Errorf("ロックファイルクローズエラー: %w", err)
	}

	// ロックファイルを削除
	if err := os.Remove(fl.filePath); err != nil {
		// ファイルが既に削除されている場合はエラーを無視
		if !os.IsNotExist(err) {
			return fmt.Errorf("ロックファイル削除エラー: %w", err)
		}
	}

	fl.file = nil
	return nil
}
