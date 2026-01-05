package redmine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/logger"
)

// Client はRedmine APIクライアント
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient は新しいAPIクライアントを作成
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchAllIssues は全チケットを取得（ページネーション対応）
// VBA版のFetchAllIssues関数に相当
//
// Redmine APIの制限により、複数チケット取得時はinclude=journalsが機能しないため、
// includeJournals=trueの場合は各チケットを個別に再取得します。
func (c *Client) FetchAllIssues(filterURL string, includeJournals bool, dateFilter *DateFilter, progress func(current, total int)) ([]*Issue, error) {
	const limit = 100
	offset := 0
	totalCount := -1
	allIssues := []*Issue{}

	logger.Section("Redmine APIからチケット取得")
	logger.Info("BaseURL: %s", c.baseURL)
	logger.Info("FilterURL: %s", filterURL)
	logger.Info("ページサイズ: %d件", limit)
	logger.Info("ジャーナル取得: %v", includeJournals)
	if dateFilter != nil {
		logger.Info("日付フィルタ: %s %s 〜 %s",
			dateFilter.Field,
			dateFilter.Start.Format("2006/01/02 15:04:05"),
			dateFilter.End.Format("2006/01/02 15:04:05"))
	}

	// Step 1: まず全チケットをjournalsなしで取得（高速）
	pageCount := 0
	for {
		pageCount++
		// URLを構築（journalsは含めない）
		requestURL := c.buildURL(filterURL, limit, offset, false, dateFilter)

		// 進捗表示
		if progress != nil {
			if totalCount > 0 {
				progress(len(allIssues), totalCount)
			} else {
				progress(len(allIssues), 0)
			}
		}

		logger.Debug("ページ%d取得中 (offset=%d)...", pageCount, offset)

		// APIリクエスト
		resp, err := c.fetch(requestURL)
		if err != nil {
			return nil, err
		}

		// 初回のみtotal_countを取得
		if totalCount == -1 {
			totalCount = resp.TotalCount
			logger.Info("合計チケット数: %d件", totalCount)
		}

		allIssues = append(allIssues, resp.Issues...)
		logger.Debug("ページ%d: %d件取得 (累計: %d/%d)", pageCount, len(resp.Issues), len(allIssues), totalCount)

		// 全件取得完了
		if len(allIssues) >= totalCount {
			break
		}

		offset += limit
	}

	logger.Info("チケット取得完了: %d件 (%dページ)", len(allIssues), pageCount)

	// Step 2: journalsが必要な場合、各チケットを個別に再取得
	// Redmine APIの制限: 複数チケット取得時はinclude=journalsが機能しない
	if includeJournals && len(allIssues) > 0 {
		logger.Section("ジャーナル（コメント）取得")
		logger.Info("各チケットを個別取得中...")
		fmt.Fprintf(os.Stderr, "[INFO] ジャーナル取得中（各チケットを個別取得）...\n")
		journalCount := 0
		errorCount := 0

		for i, issue := range allIssues {
			if progress != nil {
				progress(i+1, len(allIssues))
			}

			// 個別チケットを取得
			detailedIssue, err := c.FetchIssue(issue.ID)
			if err != nil {
				// エラーが発生してもスキップして続行
				errorCount++
				logger.Warn("Issue #%d のジャーナル取得失敗: %v", issue.ID, err)
				fmt.Fprintf(os.Stderr, "[WARN] Issue #%d のジャーナル取得失敗: %v\n", issue.ID, err)
				continue
			}

			// journalsを既存のissueにコピー
			allIssues[i].Journals = detailedIssue.Journals
			journalCount += len(detailedIssue.Journals)
		}

		logger.Info("ジャーナル取得完了: %d件のジャーナル (エラー: %d件)", journalCount, errorCount)
	}

	return allIssues, nil
}

// FetchIssue は単一のチケットをjournals付きで取得
func (c *Client) FetchIssue(issueID int) (*Issue, error) {
	url := fmt.Sprintf("%s/issues/%d.json?include=journals", c.baseURL, issueID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	req.Header.Set("X-Redmine-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// JSONパース（単一チケットのレスポンス形式）
	var result struct {
		Issue *Issue `json:"issue"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %w", err)
	}

	return result.Issue, nil
}

// buildURL はVBA版と同じロジックでURLを構築
func (c *Client) buildURL(filterURL string, limit, offset int, includeJournals bool, dateFilter *DateFilter) string {
	baseURL := c.baseURL + filterURL

	// URLにクエリパラメータが既に含まれているかチェック
	separator := "&"
	if !strings.Contains(filterURL, "?") {
		separator = "?"
	}

	url := fmt.Sprintf("%s%slimit=%d&offset=%d", baseURL, separator, limit, offset)

	// ジャーナル（コメント）を含める場合
	if includeJournals {
		url += "&include=journals"
	}

	// 日時フィルタを追加
	if dateFilter != nil {
		url += "&" + dateFilter.ToQueryString()
	}

	// デバッグ: 構築したURLを表示（APIキーは除く）
	logger.Debug("Request URL: %s (includeJournals=%v)", url, includeJournals)

	return url
}

// fetch はHTTP GETリクエストを実行
func (c *Client) fetch(url string) (*APIResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	// VBA版と同じヘッダーを設定
	req.Header.Set("X-Redmine-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	// VBA版と同じエラーハンドリング
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// JSONパース
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %w", err)
	}

	// デバッグ: レスポンスのジャーナル情報を表示
	if len(apiResp.Issues) > 0 {
		totalJournals := 0
		for _, issue := range apiResp.Issues {
			totalJournals += len(issue.Journals)
		}
		logger.Debug("API Response: %d issues, %d journals total",
			len(apiResp.Issues), totalJournals)
	}

	return &apiResp, nil
}
