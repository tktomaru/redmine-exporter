package processor

import (
	"regexp"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// Processor はチケットデータの処理を行う
type Processor struct {
	cleaningPatterns []*regexp.Regexp
}

// NewProcessor は新しいProcessorを作成
func NewProcessor(patterns []string) (*Processor, error) {
	regexps := make([]*regexp.Regexp, 0, len(patterns))

	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		// VBA版はエラーをスキップするのでここでも同様に
		re, err := regexp.Compile(pattern)
		if err != nil {
			// 不正な正規表現はスキップ（VBA版と同じ動作）
			continue
		}
		regexps = append(regexps, re)
	}

	return &Processor{
		cleaningPatterns: regexps,
	}, nil
}

// Process は全チケットを処理し、親子関係を構築
// VBA版のメインロジック（行32-48）に相当
func (p *Processor) Process(issues []*redmine.Issue) []*redmine.Issue {
	// チケットIDでマップ作成
	byID := make(map[int]*redmine.Issue)
	for _, issue := range issues {
		byID[issue.ID] = issue

		// タイトルクリーニングと要約抽出を実行
		issue.CleanedSubject = p.CleanTitle(issue.Subject)
		issue.Summary = p.ExtractSummary(issue.Description)
	}

	// 親子関係を構築
	roots := []*redmine.Issue{}
	for _, issue := range issues {
		if issue.Parent != nil {
			// 親チケットの子リストに追加
			if parent, exists := byID[issue.Parent.ID]; exists {
				parent.Children = append(parent.Children, issue)
			}
		} else {
			// 親を持たないチケットはルート
			roots = append(roots, issue)
		}
	}

	return roots
}

// CleanTitle はタイトルをクリーニング
// VBA版のCleanTitle関数（行216-246）に相当
func (p *Processor) CleanTitle(subject string) string {
	result := subject

	for _, pattern := range p.cleaningPatterns {
		result = pattern.ReplaceAllString(result, "")
	}

	return result
}

// ExtractSummary は要約を抽出
// VBA版のExtractSummary関数（行190-214）に相当
func (p *Processor) ExtractSummary(description string) string {
	const (
		startTag = "[要約]"
		endTag   = "[/要約]"
	)

	// [要約]タグの検索
	startPos := strings.Index(description, startTag)
	if startPos >= 0 {
		endPos := strings.Index(description[startPos+len(startTag):], endTag)
		if endPos >= 0 {
			summary := description[startPos+len(startTag) : startPos+len(startTag)+endPos]
			return strings.TrimSpace(summary)
		}
	}

	// タグがない場合はFirstLine処理
	return p.firstLine(description)
}

// firstLine は最初の非空行を返す
// VBA版のFirstLine関数（行174-188）に相当
func (p *Processor) firstLine(s string) string {
	// 改行を統一
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

// GetAssignee は担当者名を取得（未割り当ての場合はデフォルト値）
// VBA版のAssigneeOrUnassigned関数（行256-262）に相当
func GetAssignee(issue *redmine.Issue) string {
	if issue.AssignedTo != nil && issue.AssignedTo.Name != "" {
		return issue.AssignedTo.Name
	}
	return "担当者未定"
}
