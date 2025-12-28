package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/config"
	"github.com/tktomaru/redmine-exporter/internal/filter"
	"github.com/tktomaru/redmine-exporter/internal/formatter"
	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
	"github.com/tktomaru/redmine-exporter/internal/state"
	"github.com/tktomaru/redmine-exporter/internal/stats"
)

const version = "1.0.0"

// parseTags はタグ文字列をパースしてTagConfigの配列を返す
// commentsMaxはコメント抽出の上限（n:3の3）で、タグごとの上限として扱われる
// 形式: "要約:3,進捗:5,課題" + commentsMax=3 → [{Name: "要約", Limit: 3}, {Name: "進捗", Limit: 3}, {Name: "課題", Limit: 3}]
func parseTags(tagsStr string, commentsMax int) ([]processor.TagConfig, []string, error) {
	if tagsStr == "" {
		return nil, nil, nil
	}

	parts := strings.Split(tagsStr, ",")
	configs := make([]processor.TagConfig, 0, len(parts))
	names := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var limit int

		// "タグ名:件数" の形式をパース
		if strings.Contains(part, ":") {
			subparts := strings.SplitN(part, ":", 2)
			name := strings.TrimSpace(subparts[0])
			limitStr := strings.TrimSpace(subparts[1])

			tagLimit, err := strconv.Atoi(limitStr)
			if err != nil {
				return nil, nil, fmt.Errorf("タグ %s の件数指定が不正です: %s", name, limitStr)
			}
			if tagLimit < 0 {
				return nil, nil, fmt.Errorf("タグ %s の件数は0以上を指定してください: %d", name, tagLimit)
			}

			// タグ個別指定とコメント上限の小さい方を採用（0は無制限）
			if commentsMax > 0 && tagLimit > 0 {
				limit = min(tagLimit, commentsMax)
			} else if commentsMax > 0 {
				limit = commentsMax
			} else {
				limit = tagLimit
			}

			configs = append(configs, processor.TagConfig{Name: name, Limit: limit})
			names = append(names, name)
		} else {
			// 件数指定なし → commentsMaxを使用
			if commentsMax > 0 {
				limit = commentsMax
			} else {
				limit = 0 // 無制限
			}

			configs = append(configs, processor.TagConfig{Name: part, Limit: limit})
			names = append(names, part)
		}
	}

	return configs, names, nil
}

// min は2つの整数の最小値を返す
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseCommentsLimit はコメントモードから件数上限を抽出
// "n:3" → 3, "last" → 0, "all" → 0
func parseCommentsLimit(commentsMode string) (int, error) {
	if commentsMode == "" || commentsMode == "last" || commentsMode == "all" {
		return 0, nil
	}

	// "n:3" 形式をパース
	if strings.HasPrefix(commentsMode, "n:") {
		limitStr := strings.TrimPrefix(commentsMode, "n:")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, fmt.Errorf("コメント件数の形式エラー: %s", commentsMode)
		}
		if limit <= 0 {
			return 0, fmt.Errorf("コメント件数は1以上を指定してください: %d", limit)
		}
		return limit, nil
	}

	return 0, fmt.Errorf("不正なコメントモード: %s", commentsMode)
}

func main() {
	// コマンドライン引数の定義
	var (
		configPath      = flag.String("c", "redmine.config", "設定ファイルのパス")
		outputPath      = flag.String("o", "", "出力ファイルのパス（必須）")
		showVersion     = flag.Bool("v", false, "バージョン情報を表示")
		mode            = flag.String("mode", "", "出力モード (summary, full, tags) ※設定ファイルより優先")
		tags            = flag.String("tags", "", "抽出するタグ名（カンマ区切り、個別上限指定可） 例: 要約:5,進捗,課題:2")
		includeComments = flag.Bool("include-comments", false, "コメントからもタグを抽出する")
		tagsOrder       = flag.String("tags-order", "newest", "タグの表示順序 (newest, oldest) ※コメントから抽出されたタグの並び順")

		// 週報機能（フェーズ1）
		week      = flag.String("week", "", "週指定 (last, this, YYYY-WW) 例: last, 2025-01")
		weekStart = flag.String("week-start", "mon", "週の起点 (mon, sun)")
		dateField = flag.String("date-field", "updated_on", "日時フィールド (updated_on, created_on, start_date, due_date)")

		// コメント制御（フェーズ2）
		comments       = flag.String("comments", "", "コメント抽出モード (last, all, n:3) ※n:3はタグごとの上限にもなる")
		commentsSince  = flag.String("comments-since", "", "コメント抽出の開始日時 (auto, start, YYYY-MM-DD)")
		commentsBy     = flag.String("comments-by", "", "コメント抽出対象ユーザー")
		preferComments = flag.Bool("prefer-comments", false, "説明文よりコメントを優先")

		// グルーピング・ソート（フェーズ3）
		groupBy = flag.String("group-by", "", "グルーピング方法 (assignee, status, tracker, project, priority)")
		sortBy  = flag.String("sort", "", "ソート方法 (field または field:asc/desc, 例: updated_on, updated_on:asc, due_date:desc)")

		// State管理（フェーズ4）
		stateFile = flag.String("state", "", "Stateファイルのパス（差分運用）")
		since     = flag.String("since", "", "開始日時 (auto, YYYY-MM-DD)")
		until     = flag.String("until", "", "終了日時 (auto, YYYY-MM-DD)")

		// テンプレート機能（フェーズ5）
		templatePath = flag.String("template", "", "テンプレートファイルのパス (.tmpl)")
		stdout       = flag.Bool("stdout", false, "標準出力に出力")

		// 統計・メトリクス（フェーズ6）
		showStats      = flag.Bool("stats", false, "統計情報を表示")
		includeMetrics = flag.Bool("include-metrics", false, "詳細メトリクスを含める")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Redmine Exporter v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "使い方:\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.xlsx\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.md --mode full\n")
		fmt.Fprintf(os.Stderr, "  redmine-exporter -o output.txt --mode tags --tags \"要約,進捗,課題\" --comments n:3 --include-comments\n")
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
		fmt.Fprintf(os.Stderr, "\nタグ機能:\n")
		fmt.Fprintf(os.Stderr, "  --tags \"要約,進捗,課題\" でタグを指定\n")
		fmt.Fprintf(os.Stderr, "  --tags \"要約:3,進捗:5,課題\" でタグごとに個別の上限を指定\n")
		fmt.Fprintf(os.Stderr, "  --tags-order newest でコメントのタグを新しい順に表示（デフォルト）\n")
		fmt.Fprintf(os.Stderr, "  --tags-order oldest でコメントのタグを古い順に表示\n")
		fmt.Fprintf(os.Stderr, "  --comments n:3 がすべてのタグの共通上限（個別指定と比較して小さい方を採用）\n")
		fmt.Fprintf(os.Stderr, "  例: --comments n:3 --tags \"要約:5,進捗\" → 要約は3件、進捗は3件\n")
		fmt.Fprintf(os.Stderr, "\n週報機能:\n")
		fmt.Fprintf(os.Stderr, "  --week last で先週分のチケットを一発で取得\n")
		fmt.Fprintf(os.Stderr, "  --week-start で週の起点を月曜/日曜で切り替え\n")
		fmt.Fprintf(os.Stderr, "  --date-field で更新日時/作成日時などでフィルタ\n")
		fmt.Fprintf(os.Stderr, "\nコメント制御:\n")
		fmt.Fprintf(os.Stderr, "  --comments last で最新コメントのみ抽出\n")
		fmt.Fprintf(os.Stderr, "  --comments n:3 で最新3件のコメントを抽出\n")
		fmt.Fprintf(os.Stderr, "  --comments-since auto/start で週の開始以降のコメントのみ\n")
		fmt.Fprintf(os.Stderr, "  --comments-by で特定ユーザーのコメントのみ抽出\n")
		fmt.Fprintf(os.Stderr, "\nグルーピング・ソート:\n")
		fmt.Fprintf(os.Stderr, "  --group-by assignee で担当者別にグルーピング\n")
		fmt.Fprintf(os.Stderr, "  --group-by status でステータス別にグルーピング\n")
		fmt.Fprintf(os.Stderr, "  --sort updated_on で更新日時順にソート（デフォルト：降順）\n")
		fmt.Fprintf(os.Stderr, "  --sort updated_on:asc で昇順、updated_on:desc で降順\n")
		fmt.Fprintf(os.Stderr, "  --sort due_date で期日順にソート（デフォルト：昇順）\n")
		fmt.Fprintf(os.Stderr, "  対応フィールド: updated_on, created_on, due_date, start_date, priority, id\n")
		fmt.Fprintf(os.Stderr, "\n差分運用（State管理）:\n")
		fmt.Fprintf(os.Stderr, "  --state .state.json でState管理を有効化\n")
		fmt.Fprintf(os.Stderr, "  --since auto で前回実行以降のチケットのみ取得\n")
		fmt.Fprintf(os.Stderr, "  --until auto で現在時刻までのチケットを取得\n")
		fmt.Fprintf(os.Stderr, "\nテンプレート機能:\n")
		fmt.Fprintf(os.Stderr, "  --template weekly.tmpl でカスタムテンプレートを使用\n")
		fmt.Fprintf(os.Stderr, "  --stdout で標準出力に出力（ファイル作成なし）\n")
		fmt.Fprintf(os.Stderr, "\n統計・メトリクス:\n")
		fmt.Fprintf(os.Stderr, "  --stats で統計情報を表示（総件数、ステータス別など）\n")
		fmt.Fprintf(os.Stderr, "  --include-metrics で詳細メトリクスを含める（期限切れ、コメント統計など）\n")
	}

	flag.Parse()

	// バージョン表示
	if *showVersion {
		fmt.Printf("Redmine Exporter v%s\n", version)
		os.Exit(0)
	}

	// 出力パスのチェック（stdoutモードでない場合のみ）
	if *outputPath == "" && !*stdout {
		fmt.Fprintln(os.Stderr, "エラー: 出力ファイルを指定してください (-o) または --stdout を使用してください")
		flag.Usage()
		os.Exit(1)
	}

	// 実行
	if err := run(*configPath, *outputPath, *mode, *tags, *includeComments, *tagsOrder, *week, *weekStart, *dateField, *comments, *commentsSince, *commentsBy, *preferComments, *groupBy, *sortBy, *stateFile, *since, *until, *templatePath, *stdout, *showStats, *includeMetrics); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath, outputPath, modeFlag, tagsFlag string, includeCommentsFlag bool, tagsOrderFlag, weekFlag, weekStartFlag, dateFieldFlag, commentsMode, commentsSinceFlag, commentsByFlag string, preferCommentsFlag bool, groupByFlag, sortByFlag, stateFileFlag, sinceFlag, untilFlag, templatePathFlag string, stdoutFlag, showStatsFlag, includeMetricsFlag bool) error {
	// 0. State管理の初期化（指定されている場合）
	var stateMgr *state.Manager
	var stateData *state.State
	var fileLock *state.FileLock

	// 統計計算用の期間（週報機能や差分運用で設定される）
	var statsWeekStart, statsWeekEnd time.Time

	if stateFileFlag != "" {
		// ファイルロック取得
		lock, err := state.AcquireLock(stateFileFlag, 10*time.Second)
		if err != nil {
			return fmt.Errorf("ファイルロック取得エラー: %w", err)
		}
		fileLock = lock
		defer fileLock.Release()

		// State読み込み
		stateMgr = state.NewManager(stateFileFlag)
		stateData, err = stateMgr.Load()
		if err != nil {
			// State破損の場合は警告を表示
			fmt.Fprintf(os.Stderr, "警告: %v\n", err)
		}

		// 実行開始時刻を記録
		stateMgr.UpdateLastRun(stateData)
	}

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

	// コメント件数の上限を取得
	commentsMax := 0
	if commentsMode != "" {
		var err error
		commentsMax, err = parseCommentsLimit(commentsMode)
		if err != nil {
			return fmt.Errorf("コメント設定のパースエラー: %w", err)
		}
	}

	// タグのパース（件数制限をサポート、commentsが上限）
	var tagConfigs []processor.TagConfig
	if tagsFlag != "" {
		var err error
		tagConfigs, cfg.Output.TagNames, err = parseTags(tagsFlag, commentsMax)
		if err != nil {
			return fmt.Errorf("タグのパースエラー: %w", err)
		}
	} else if len(cfg.Output.TagNames) > 0 {
		// 設定ファイルから読み込んだ場合はcommentsMaxを適用
		tagConfigs = make([]processor.TagConfig, len(cfg.Output.TagNames))
		for i, name := range cfg.Output.TagNames {
			tagConfigs[i] = processor.TagConfig{Name: name, Limit: commentsMax}
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

		// 統計計算用の期間を設定
		statsWeekStart = start
		statsWeekEnd = end

		fmt.Printf("期間フィルタ: %s %s 〜 %s\n", dateFieldFlag, start.Format("2006/01/02"), end.Format("2006/01/02"))
	}

	// since/untilフラグの処理（State管理との連携）
	if sinceFlag != "" || untilFlag != "" {
		var start, end time.Time

		// since処理
		if sinceFlag == "auto" {
			if stateData != nil && !stateData.LastSuccessRun.IsZero() {
				start = stateData.LastSuccessRun
				fmt.Printf("差分運用: 前回成功実行 %s 以降のチケットを取得\n", start.Format("2006/01/02 15:04:05"))
			} else {
				return fmt.Errorf("--since auto を使用するには --state でStateファイルを指定し、過去に成功実行が必要です")
			}
		} else if sinceFlag != "" {
			var err error
			start, err = time.Parse("2006-01-02", sinceFlag)
			if err != nil {
				return fmt.Errorf("--since の日付形式エラー: %w", err)
			}
		} else if dateFilter != nil {
			start = dateFilter.Start
		}

		// until処理
		if untilFlag == "auto" {
			end = time.Now()
		} else if untilFlag != "" {
			var err error
			end, err = time.Parse("2006-01-02", untilFlag)
			if err != nil {
				return fmt.Errorf("--until の日付形式エラー: %w", err)
			}
			// 終了日を23:59:59に設定
			end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
		} else if dateFilter != nil {
			end = dateFilter.End
		} else {
			end = time.Now()
		}

		// DateFilterを作成/更新
		dateFilter = &redmine.DateFilter{
			Field: dateFieldFlag,
			Start: start,
			End:   end,
		}

		// 統計計算用の期間を設定
		statsWeekStart = start
		statsWeekEnd = end

		fmt.Printf("期間フィルタ: %s %s 〜 %s\n", dateFieldFlag, start.Format("2006/01/02 15:04:05"), end.Format("2006/01/02 15:04:05"))
	}

	// 2. Redmine APIクライアント作成
	client := redmine.NewClient(cfg.Redmine.BaseURL, cfg.Redmine.APIKey)

	// 3. 全チケット取得（進捗表示付き）
	// コメント関連の機能を使用する場合は、必ずjournalsを取得
	needsJournals := cfg.Output.IncludeComments ||
		commentsMode != "" ||
		commentsSinceFlag != "" ||
		commentsByFlag != "" ||
		preferCommentsFlag

	// デバッグ情報
	fmt.Fprintf(os.Stderr, "[DEBUG] needsJournals=%v (IncludeComments=%v, mode=%s)\n",
		needsJournals, cfg.Output.IncludeComments, commentsMode)

	fmt.Println("Redmineからチケットを取得中...")
	issues, err := client.FetchAllIssues(cfg.Redmine.FilterURL, needsJournals, dateFilter, func(current, total int) {
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

	// デバッグ：ジャーナル情報を表示
	if needsJournals {
		totalJournals := 0
		journalsWithNotes := 0
		for _, issue := range issues {
			totalJournals += len(issue.Journals)
			for _, j := range issue.Journals {
				if j.Notes != "" {
					journalsWithNotes++
				}
			}
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] 取得したジャーナル: 合計%d件 (Notes有り: %d件)\n",
			totalJournals, journalsWithNotes)
	}

	// 3.5. コメントフィルタの適用
	if commentsMode != "" || commentsSinceFlag != "" || commentsByFlag != "" {
		fmt.Println("コメントをフィルタリング中...")

		// commentsSinceの解釈（"auto" または "start" の場合は週の開始日を使用）
		var commentsSinceDate *time.Time
		if (commentsSinceFlag == "auto" || commentsSinceFlag == "start") && dateFilter != nil {
			commentsSinceDate = &dateFilter.Start
		} else if commentsSinceFlag != "" && commentsSinceFlag != "auto" && commentsSinceFlag != "start" {
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
	proc, err := processor.NewProcessor(cfg.TitleCleaning.Patterns, tagConfigs, cfg.Output.Mode, preferCommentsFlag, cfg.Output.IncludeComments, tagsOrderFlag)
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
	// stdoutモードの場合、outputPathが空の可能性があるため、テンプレートパスまたはデフォルトを使用
	formatterOutputPath := outputPath
	if stdoutFlag && formatterOutputPath == "" {
		// stdoutモードでoutputPathが空の場合、拡張子判定用にダミーパス
		if templatePathFlag != "" {
			formatterOutputPath = templatePathFlag
		} else {
			formatterOutputPath = "stdout.md" // デフォルトはMarkdown
		}
	}

	fmtr, err := formatter.DetectFormatter(formatterOutputPath, cfg.Output.Mode, cfg.Output.TagNames, templatePathFlag)
	if err != nil {
		return err
	}

	// 5.5. 統計計算（--stats または --include-metrics が指定されている場合）
	if showStatsFlag || includeMetricsFlag {
		// 統計期間が設定されていない場合は、デフォルト期間を使用
		if statsWeekStart.IsZero() {
			statsWeekStart = time.Now().AddDate(0, 0, -7) // 過去7日間
		}
		if statsWeekEnd.IsZero() {
			statsWeekEnd = time.Now()
		}

		// 統計を計算
		weeklyStats := stats.Calculate(roots, statsWeekStart, statsWeekEnd)

		// テンプレートフォーマッターの場合は統計を設定
		if tmplFmtr, ok := fmtr.(*formatter.TemplateFormatter); ok {
			tmplFmtr.SetStats(weeklyStats, statsWeekStart, statsWeekEnd)
		}

		// --stats フラグが指定されている場合は、標準エラー出力に統計を表示
		if showStatsFlag {
			fmt.Fprintf(os.Stderr, "\n=== 統計情報 ===\n")
			fmt.Fprintf(os.Stderr, "総チケット数: %d\n", weeklyStats.TotalIssues)
			fmt.Fprintf(os.Stderr, "\nステータス別:\n")
			for status, count := range weeklyStats.ByStatus {
				fmt.Fprintf(os.Stderr, "  %s: %d\n", status, count)
			}
			fmt.Fprintf(os.Stderr, "\n担当者別:\n")
			for assignee, count := range weeklyStats.ByAssignee {
				fmt.Fprintf(os.Stderr, "  %s: %d\n", assignee, count)
			}
		}

		// --include-metrics フラグが指定されている場合は詳細メトリクスを表示
		if includeMetricsFlag {
			fmt.Fprintf(os.Stderr, "\n=== 詳細メトリクス ===\n")
			fmt.Fprintf(os.Stderr, "新規作成: %d\n", weeklyStats.NewIssues)
			fmt.Fprintf(os.Stderr, "更新: %d\n", weeklyStats.UpdatedIssues)
			fmt.Fprintf(os.Stderr, "完了: %d\n", weeklyStats.ClosedIssues)
			fmt.Fprintf(os.Stderr, "期限切れ: %d\n", len(weeklyStats.OverdueTasks))
			fmt.Fprintf(os.Stderr, "期限間近（7日以内）: %d\n", len(weeklyStats.DueSoonTasks))
			fmt.Fprintf(os.Stderr, "\nコメント統計:\n")
			fmt.Fprintf(os.Stderr, "  総コメント数: %d\n", weeklyStats.CommentStats.TotalComments)
			fmt.Fprintf(os.Stderr, "  コメントのあるチケット数: %d\n", weeklyStats.CommentStats.IssuesWithComments)
		}
	}

	// 6. 出力
	if stdoutFlag {
		// 標準出力に出力
		fmt.Fprintln(os.Stderr, "標準出力に出力中...")
		if err := fmtr.Format(roots, os.Stdout); err != nil {
			return fmt.Errorf("出力エラー: %w", err)
		}
		fmt.Fprintf(os.Stderr, "出力完了: %d 件のチケット\n", ticketCount)
	} else {
		// ファイルに出力
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
	}

	// 7. State保存（成功時のみ）
	if stateMgr != nil && stateData != nil {
		stateMgr.UpdateLastSuccessRun(stateData)
		stateData.Version = version

		// フィルタ設定を記録
		if weekFlag != "" {
			stateMgr.SetFilterConfig(stateData, "week", weekFlag)
		}
		if dateFieldFlag != "" {
			stateMgr.SetFilterConfig(stateData, "date_field", dateFieldFlag)
		}

		if err := stateMgr.Save(stateData); err != nil {
			fmt.Fprintf(os.Stderr, "警告: State保存エラー: %v\n", err)
		} else {
			fmt.Println("State保存完了")
		}
	}

	return nil
}
