package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/config"
	"github.com/tktomaru/redmine-exporter/internal/formatter"
	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

const version = "1.0.0"

func main() {
	// コマンドライン引数の定義
	var (
		configPath      = flag.String("c", "redmine.config", "設定ファイルのパス")
		outputPath      = flag.String("o", "", "出力ファイルのパス（必須）")
		showVersion     = flag.Bool("v", false, "バージョン情報を表示")
		mode            = flag.String("mode", "", "出力モード (summary, full, tags) ※設定ファイルより優先")
		tags            = flag.String("tags", "", "抽出するタグ名（カンマ区切り） 例: 要約,進捗,課題")
		includeComments = flag.Bool("include-comments", false, "コメントからもタグを抽出する")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Redmine Exporter v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "使い方:\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.xlsx\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.md --mode full\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.txt --mode tags --tags \"要約,進捗\" --include-comments\n\n")
		fmt.Fprintf(os.Stderr, "オプション:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n対応する出力形式:\n")
		fmt.Fprintf(os.Stderr, "  .md   - Markdown形式\n")
		fmt.Fprintf(os.Stderr, "  .txt  - テキスト形式\n")
		fmt.Fprintf(os.Stderr, "  .xlsx - Excel形式\n")
		fmt.Fprintf(os.Stderr, "\n出力モード:\n")
		fmt.Fprintf(os.Stderr, "  summary - 要約のみ出力（デフォルト）\n")
		fmt.Fprintf(os.Stderr, "  full    - すべてのフィールドを出力\n")
		fmt.Fprintf(os.Stderr, "  tags    - 指定したタグの内容を抽出\n")
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
	if err := run(*configPath, *outputPath, *mode, *tags, *includeComments); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath, outputPath, modeFlag, tagsFlag string, includeCommentsFlag bool) error {
	// 1. 設定ファイル読み込み
	fmt.Printf("設定ファイルを読み込んでいます: %s\n", configPath)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	// コマンドラインフラグで設定を上書き
	if modeFlag != "" {
		cfg.Output.Mode = modeFlag
	}
	if tagsFlag != "" {
		cfg.Output.TagNames = strings.Split(tagsFlag, ",")
		for i := range cfg.Output.TagNames {
			cfg.Output.TagNames[i] = strings.TrimSpace(cfg.Output.TagNames[i])
		}
	}
	if includeCommentsFlag {
		cfg.Output.IncludeComments = true
	}

	// 2. Redmine APIクライアント作成
	client := redmine.NewClient(cfg.Redmine.BaseURL, cfg.Redmine.APIKey)

	// 3. 全チケット取得（進捗表示付き）
	fmt.Println("Redmineからチケットを取得中...")
	issues, err := client.FetchAllIssues(cfg.Redmine.FilterURL, cfg.Output.IncludeComments, func(current, total int) {
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
	proc, err := processor.NewProcessor(cfg.TitleCleaning.Patterns, cfg.Output.TagNames, cfg.Output.Mode)
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
