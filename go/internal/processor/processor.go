package processor

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// Processor はチケットデータの処理を行う
type Processor struct {
	cleaningPatterns []*regexp.Regexp
	tagNames         []string
	mode             string
	preferComments   bool // 説明文よりコメントを優先
	includeComments  bool // コメントからもタグを抽出
}

// NewProcessor は新しいProcessorを作成
func NewProcessor(patterns []string, tagNames []string, mode string, preferComments bool, includeComments bool) (*Processor, error) {
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
		tagNames:         tagNames,
		mode:             mode,
		preferComments:   preferComments,
		includeComments:  includeComments,
	}, nil
}

// Process は全チケットを処理し、親子関係を構築
// VBA版のメインロジック（行32-48）に相当
func (p *Processor) Process(issues []*redmine.Issue) []*redmine.Issue {
	// チケットIDでマップ作成
	byID := make(map[int]*redmine.Issue)
	for _, issue := range issues {
		byID[issue.ID] = issue

		// タイトルクリーニング
		issue.CleanedSubject = p.CleanTitle(issue.Subject)

		// prefer-commentsの場合、コメントから内容を取得
		contentSource := issue.Description
		if p.preferComments && len(issue.Journals) > 0 {
			// 最新のコメント（Notesが空でないもの）を取得
			for i := len(issue.Journals) - 1; i >= 0; i-- {
				if issue.Journals[i].Notes != "" {
					contentSource = issue.Journals[i].Notes
					break
				}
			}
		}

		// モードに応じて処理
		switch p.mode {
		case "tags":
			// タグ抽出モード：複数のタグを抽出
			issue.ExtractedTags = p.ExtractTags(p.tagNames, contentSource, issue.Journals)
			// デバッグ：タグ抽出結果を表示
			if len(issue.ExtractedTags) > 0 {
				// fmt.Fprintf(os.Stderr, "[DEBUG] Issue #%d: ExtractedTags=%v (journals=%d)\n",
				// 	issue.ID, issue.ExtractedTags, len(issue.Journals))
			}
			// 後方互換性のため、要約タグがあればSummaryにも設定（最初の値を使用）
			if summaries, ok := issue.ExtractedTags["要約"]; ok && len(summaries) > 0 {
				issue.Summary = summaries[0]
			}
		case "full":
			// フルモード：すべての情報を保持（特別な処理なし）
			issue.Summary = p.ExtractSummary(contentSource)
		default:
			// summaryモード：要約のみ抽出（デフォルト動作）
			issue.Summary = p.ExtractSummary(contentSource)
		}
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

// ExtractSummary は要約を抽出（後方互換性のため残す）
// VBA版のExtractSummary関数（行190-214）に相当
func (p *Processor) ExtractSummary(description string) string {
	return p.ExtractTag("要約", description)
}

// ExtractTag は指定されたタグの内容を抽出
func (p *Processor) ExtractTag(tagName, text string) string {
	startTag := "[" + tagName + "]"
	endTag := "[/" + tagName + "]"

	// タグの検索
	startPos := strings.Index(text, startTag)
	if startPos >= 0 {
		endPos := strings.Index(text[startPos+len(startTag):], endTag)
		if endPos >= 0 {
			content := text[startPos+len(startTag) : startPos+len(startTag)+endPos]
			return strings.TrimSpace(content)
		}
	}

	// タグがない場合は空文字列を返す（要約タグの場合のみFirstLineを返す）
	// if tagName == "要約" {
	// 	return p.firstLine(text)
	// }
	return ""
}

// ExtractTags は複数のタグを抽出してマップで返す
// 各タグは複数の値を配列で保持する
func (p *Processor) ExtractTags(tagNames []string, description string, journals []redmine.Journal) map[string][]string {
	result := make(map[string][]string)

	// 説明文から抽出
	for _, tagName := range tagNames {
		if content := p.ExtractTag(tagName, description); content != "" {
			result[tagName] = append(result[tagName], content)
			// デバッグ: 説明文から抽出
			fmt.Fprintf(os.Stderr, "[DEBUG] 説明文から抽出: タグ=%s, 内容=%q\n", tagName, content)
		}
	}

	// ジャーナル（コメント）からも抽出（includeCommentsがtrueの場合のみ）
	if p.includeComments {
		fmt.Fprintf(os.Stderr, "[DEBUG] ExtractTags: includeComments=true, journals=%d\n", len(journals))
		// 最新のコメントから順に処理（逆順）
		// for i := len(journals) - 1; i >= 0; i-- {
		for i := 0; i < len(journals); i++ {
			journal := journals[i]
			if journal.Notes == "" {
				continue
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] ジャーナル[%d]: Notes=%q\n", i, journal.Notes)
			for _, tagName := range tagNames {
				if content := p.ExtractTag(tagName, journal.Notes); content != "" {
					result[tagName] = append(result[tagName], content)
					fmt.Fprintf(os.Stderr, "[DEBUG] ジャーナルから抽出成功: タグ=%s, 内容=%q (合計%d個)\n", tagName, content, len(result[tagName]))
				}
			}
		}
	}

	return result
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
