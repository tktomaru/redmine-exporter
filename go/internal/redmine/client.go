package redmine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
func (c *Client) FetchAllIssues(filterURL string, includeJournals bool, dateFilter *DateFilter, progress func(current, total int)) ([]*Issue, error) {
	const limit = 100
	offset := 0
	totalCount := -1
	allIssues := []*Issue{}

	for {
		// URLを構築（VBA版と同じロジック）
		requestURL := c.buildURL(filterURL, limit, offset, includeJournals, dateFilter)

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

	return allIssues, nil
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

	return &apiResp, nil
}
