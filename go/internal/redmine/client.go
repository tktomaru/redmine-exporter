package redmine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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

	// Step 1: まず全チケットをjournalsなしで取得（高速）
	for {
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

		// APIリクエスト
		resp, err := c.fetch(requestURL)
		if err != nil {
			return nil, err
		}

		// 初回のみtotal_countを取得
		if totalCount == -1 {
			totalCount = resp.TotalCount
		}

		allIssues = append(allIssues, resp.Issues...)

		// 全件取得完了
		if len(allIssues) >= totalCount {
			break
		}

		offset += limit
	}

	// Step 2: journalsが必要な場合、各チケットを個別に再取得
	// Redmine APIの制限: 複数チケット取得時はinclude=journalsが機能しない
	if includeJournals && len(allIssues) > 0 {
		fmt.Fprintf(os.Stderr, "[INFO] ジャーナル取得中（各チケットを個別取得）...\n")
		for i, issue := range allIssues {
			if progress != nil {
				progress(i+1, len(allIssues))
			}

			// 個別チケットを取得
			detailedIssue, err := c.FetchIssue(issue.ID)
			if err != nil {
				// エラーが発生してもスキップして続行
				fmt.Fprintf(os.Stderr, "[WARN] Issue #%d のジャーナル取得失敗: %v\n", issue.ID, err)
				continue
			}

			// journalsを既存のissueにコピー
			allIssues[i].Journals = detailedIssue.Journals
		}
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
	fmt.Fprintf(os.Stderr, "[DEBUG] Request URL: %s (includeJournals=%v)\n", url, includeJournals)

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
		fmt.Fprintf(os.Stderr, "[DEBUG] API Response: %d issues, %d journals total\n",
			len(apiResp.Issues), totalJournals)
	}

	return &apiResp, nil
}
