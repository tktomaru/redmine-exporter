package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tktomaru/redmine-exporter/internal/config"
	"github.com/tktomaru/redmine-exporter/internal/formatter"
	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

const version = "1.0.0"

func main() {
	// コマンドライン引数の定義
	var (
		configPath  = flag.String("c", "redmine.config", "設定ファイルのパス")
		outputPath  = flag.String("o", "", "出力ファイルのパス（必須）")
		showVersion = flag.Bool("v", false, "バージョン情報を表示")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Redmine Exporter v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "使い方:\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.xlsx\n\n")
		fmt.Fprintf(os.Stderr, "オプション:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n対応する出力形式:\n")
		fmt.Fprintf(os.Stderr, "  .md   - Markdown形式\n")
		fmt.Fprintf(os.Stderr, "  .txt  - テキスト形式\n")
		fmt.Fprintf(os.Stderr, "  .xlsx - Excel形式\n")
	}

	flag.Parse()

	// バージョン表示
	if *showVersion {
		fmt.Printf("Redmine Exporter v%s\n", version)
		os.Exit(0)
	}

	// 出力パスのチェック
	if *outputPath == "" {
		fmt.Fprintln(os.Stderr, "エラー: 出力ファイルを指定してください (-o)")
		flag.Usage()
		os.Exit(1)
	}

	// 実行
	if err := run(*configPath, *outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath, outputPath string) error {
	// 1. 設定ファイル読み込み
	fmt.Printf("設定ファイルを読み込んでいます: %s\n", configPath)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	// 2. Redmine APIクライアント作成
	client := redmine.NewClient(cfg.Redmine.BaseURL, cfg.Redmine.APIKey)

	// 3. 全チケット取得（進捗表示付き）
	fmt.Println("Redmineからチケットを取得中...")
	issues, err := client.FetchAllIssues(cfg.Redmine.FilterURL, func(current, total int) {
		if total > 0 {
			fmt.Printf("\r取得中... (%d / %d)", current, total)
		} else {
			fmt.Printf("\r取得中... (%d)", current)
		}
	})
	if err != nil {
		return fmt.Errorf("チケット取得エラー: %w", err)
	}
	fmt.Printf("\r取得完了: %d 件のチケット\n", len(issues))

	// 4. データ処理
	fmt.Println("チケットを処理中...")
	proc, err := processor.NewProcessor(cfg.TitleCleaning.Patterns)
	if err != nil {
		return fmt.Errorf("プロセッサー初期化エラー: %w", err)
	}
	roots := proc.Process(issues)

	// 出力するチケット数をカウント
	ticketCount := 0
	for _, root := range roots {
		// スタンドアロンチケット（子を持たない）も1件としてカウント
		if len(root.Children) == 0 {
			ticketCount++
		} else {
			ticketCount += len(root.Children)
		}
	}

	if ticketCount == 0 {
		fmt.Println("出力するチケットがありません")
		return nil
	}

	// 5. フォーマッター選択
	fmtr, err := formatter.DetectFormatter(outputPath)
	if err != nil {
		return err
	}

	// 6. ファイル出力
	fmt.Printf("ファイルに出力中: %s\n", outputPath)

	// 出力ディレクトリが存在しない場合は作成
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ディレクトリ作成エラー: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ファイル作成エラー: %w", err)
	}
	defer file.Close()

	if err := fmtr.Format(roots, file); err != nil {
		return fmt.Errorf("出力エラー: %w", err)
	}

	fmt.Printf("出力完了: %d 件のチケット\n", ticketCount)
	return nil
}
