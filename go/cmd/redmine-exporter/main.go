package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/config"
	"github.com/tktomaru/redmine-exporter/internal/filter"
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

		// 週報機能（フェーズ1）
		week      = flag.String("week", "", "週指定 (last, this, YYYY-WW) 例: last, 2025-01")
		weekStart = flag.String("week-start", "mon", "週の起点 (mon, sun)")
		dateField = flag.String("date-field", "updated_on", "日時フィールド (updated_on, created_on, start_date, due_date)")

		// コメント制御（フェーズ2）
		comments       = flag.String("comments", "", "コメント抽出モード (last, all, n:3)")
		commentsSince  = flag.String("comments-since", "", "コメント抽出の開始日時 (start, YYYY-MM-DD)")
		commentsBy     = flag.String("comments-by", "", "コメント抽出対象ユーザー")
		preferComments = flag.Bool("prefer-comments", false, "説明文よりコメントを優先")

		// グルーピング・ソート（フェーズ3）
		groupBy = flag.String("group-by", "", "グルーピング方法 (assignee, status, tracker, project, priority)")
		sortBy  = flag.String("sort", "", "ソート方法 (updated_on, created_on, due_date, start_date, priority, id)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Redmine Exporter v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "使い方:\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.xlsx\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.md --mode full\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.txt --mode tags --tags \"要約,進捗\" --include-comments\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o weekly.md --week last --week-start mon\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o weekly.md --week last --comments last --comments-since start\n\n")
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
		fmt.Fprintf(os.Stderr, "\n週報機能:\n")
		fmt.Fprintf(os.Stderr, "  --week last で先週分のチケットを一発で取得\n")
		fmt.Fprintf(os.Stderr, "  --week-start で週の起点を月曜/日曜で切り替え\n")
		fmt.Fprintf(os.Stderr, "  --date-field で更新日時/作成日時などでフィルタ\n")
		fmt.Fprintf(os.Stderr, "\nコメント制御:\n")
		fmt.Fprintf(os.Stderr, "  --comments last で最新コメントのみ抽出\n")
		fmt.Fprintf(os.Stderr, "  --comments n:3 で最新3件のコメントを抽出\n")
		fmt.Fprintf(os.Stderr, "  --comments-since start で週の開始以降のコメントのみ\n")
		fmt.Fprintf(os.Stderr, "  --comments-by で特定ユーザーのコメントのみ抽出\n")
		fmt.Fprintf(os.Stderr, "\nグルーピング・ソート:\n")
		fmt.Fprintf(os.Stderr, "  --group-by assignee で担当者別にグルーピング\n")
		fmt.Fprintf(os.Stderr, "  --group-by status でステータス別にグルーピング\n")
		fmt.Fprintf(os.Stderr, "  --sort updated_on で更新日時順にソート\n")
		fmt.Fprintf(os.Stderr, "  --sort due_date で期日順にソート\n")
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
	if err := run(*configPath, *outputPath, *mode, *tags, *includeComments, *week, *weekStart, *dateField, *comments, *commentsSince, *commentsBy, *preferComments, *groupBy, *sortBy); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath, outputPath, modeFlag, tagsFlag string, includeCommentsFlag bool, weekFlag, weekStartFlag, dateFieldFlag, commentsMode, commentsSinceFlag, commentsByFlag string, preferCommentsFlag bool, groupByFlag, sortByFlag string) error {
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

	// 週報フィルタの構築
	var dateFilter *redmine.DateFilter
	if weekFlag != "" {
		// WeekCalculatorを作成
		wc, err := filter.NewWeekCalculator(weekStartFlag, "Asia/Tokyo")
		if err != nil {
			return fmt.Errorf("週計算エラー: %w", err)
		}

		// 週の期間を取得
		start, end, err := wc.GetWeekRange(weekFlag)
		if err != nil {
			return fmt.Errorf("週範囲計算エラー: %w", err)
		}

		// DateFilterを構築
		dateFilter = &redmine.DateFilter{
			Field: dateFieldFlag,
			Start: start,
			End:   end,
		}

		fmt.Printf("期間フィルタ: %s %s 〜 %s\n", dateFieldFlag, start.Format("2006/01/02"), end.Format("2006/01/02"))
	}

	// 2. Redmine APIクライアント作成
	client := redmine.NewClient(cfg.Redmine.BaseURL, cfg.Redmine.APIKey)

	// 3. 全チケット取得（進捗表示付き）
	fmt.Println("Redmineからチケットを取得中...")
	issues, err := client.FetchAllIssues(cfg.Redmine.FilterURL, cfg.Output.IncludeComments, dateFilter, func(current, total int) {
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

	// 3.5. コメントフィルタの適用
	if commentsMode != "" || commentsSinceFlag != "" || commentsByFlag != "" {
		fmt.Println("コメントをフィルタリング中...")

		// commentsSinceの解釈（"start" の場合は週の開始日を使用）
		var commentsSinceDate *time.Time
		if commentsSinceFlag == "start" && dateFilter != nil {
			commentsSinceDate = &dateFilter.Start
		} else if commentsSinceFlag != "" && commentsSinceFlag != "start" {
			// YYYY-MM-DD形式をパース
			t, err := time.Parse("2006-01-02", commentsSinceFlag)
			if err != nil {
				return fmt.Errorf("コメント開始日時の解析エラー: %w", err)
			}
			commentsSinceDate = &t
		}

		// CommentFilterを作成
		commentFilter, err := filter.NewCommentFilter(commentsMode, commentsSinceDate, commentsByFlag)
		if err != nil {
			return fmt.Errorf("コメントフィルタ作成エラー: %w", err)
		}

		// 各チケットのジャーナルをフィルタリング
		for _, issue := range issues {
			issue.Journals = commentFilter.Filter(issue.Journals)
		}

		fmt.Println("コメントフィルタリング完了")
	}

	// 4. データ処理
	fmt.Println("チケットを処理中...")
	proc, err := processor.NewProcessor(cfg.TitleCleaning.Patterns, cfg.Output.TagNames, cfg.Output.Mode, preferCommentsFlag)
	if err != nil {
		return fmt.Errorf("プロセッサー初期化エラー: %w", err)
	}
	roots := proc.Process(issues)

	// 4.5. グルーピング・ソート
	if sortByFlag != "" || groupByFlag != "" {
		fmt.Println("チケットをソート・グルーピング中...")

		// ルートチケットと子チケットをフラットなリストに展開
		var allIssues []*redmine.Issue
		for _, root := range roots {
			allIssues = append(allIssues, root)
			allIssues = append(allIssues, root.Children...)
		}

		// ソート
		if sortByFlag != "" {
			sorter := processor.NewSorter(sortByFlag)
			if sorter != nil {
				sorter.Sort(allIssues)
			}
		}

		// グルーピング
		if groupByFlag != "" {
			grouper := processor.NewGrouper(groupByFlag)
			if grouper != nil {
				grouped := grouper.Group(allIssues)

				// グルーピング後、各グループ内でもソートを適用
				if sortByFlag != "" {
					sorter := processor.NewSorter(sortByFlag)
					if sorter != nil {
						for _, key := range grouped.Keys {
							sorter.Sort(grouped.Groups[key])
						}
					}
				}

				// グルーピング結果をフラットに戻す
				allIssues = processor.FlattenGroupedIssues(grouped)
			}
		}

		// ソート・グルーピング後、親子関係を再構築せずにフラットなリストとして扱う
		// （グルーピング表示では親子関係より、グループ内の並びが重要）
		roots = make([]*redmine.Issue, 0, len(allIssues))
		for _, issue := range allIssues {
			// 子チケットのChildrenをクリアして、フラットに扱う
			issue.Children = nil
			roots = append(roots, issue)
		}
	}

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
	fmtr, err := formatter.DetectFormatter(outputPath, cfg.Output.Mode, cfg.Output.TagNames)
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
