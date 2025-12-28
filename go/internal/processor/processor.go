package processor

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// TagConfig はタグ名と出力件数制限を保持する
type TagConfig struct {
	Name  string // タグ名
	Limit int    // 出力件数制限（0は無制限）
}

// Processor はチケットデータの処理を行う
type Processor struct {
	cleaningPatterns []*regexp.Regexp
	tagConfigs       []TagConfig // タグ設定（名前と件数制限）
	mode             string
	preferComments   bool // 説明文よりコメントを優先
	includeComments  bool // コメントからもタグを抽出
}

// NewProcessor は新しいProcessorを作成
func NewProcessor(patterns []string, tagConfigs []TagConfig, mode string, preferComments bool, includeComments bool) (*Processor, error) {
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
		tagConfigs:       tagConfigs,
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

		// モードに応じて処理
		switch p.mode {
		case "tags":
			// タグ抽出モード：複数のタグを抽出
			// 説明文は常にissue.Descriptionから、コメントは別途独立して処理
			issue.ExtractedTags = p.ExtractTags(issue.Description, issue.Journals)
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
			issue.Summary = p.ExtractSummary(issue.Description)
		default:
			// summaryモード：要約のみ抽出（デフォルト動作）
			issue.Summary = p.ExtractSummary(issue.Description)
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
// 各タグは複数の値を配列で保持し、TagConfigの件数制限を適用する
// 説明文とコメントは独立して処理される
func (p *Processor) ExtractTags(description string, journals []redmine.Journal) map[string][]string {
	result := make(map[string][]string)

	// 説明文から抽出（説明文のみから）
	for _, tagConfig := range p.tagConfigs {
		if content := p.ExtractTag(tagConfig.Name, description); content != "" {
			result[tagConfig.Name] = append(result[tagConfig.Name], content)
			// デバッグ: 説明文から抽出
			fmt.Fprintf(os.Stderr, "[DEBUG] 説明文から抽出: タグ=%s, 内容=%q\n", tagConfig.Name, content)
		}
	}

	// ジャーナル（コメント）から抽出（includeCommentsがtrueの場合のみ）
	// コメントからの抽出は説明文の結果とは独立して行われる
	if p.includeComments {
		fmt.Fprintf(os.Stderr, "[DEBUG] ExtractTags: includeComments=true, journals=%d\n", len(journals))
		// 最新のコメントから順に処理（逆順）
		for i := len(journals) - 1; i >= 0; i-- {
			journal := journals[i]
			if journal.Notes == "" {
				continue
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] ジャーナル[%d]: Notes=%q\n", i, journal.Notes)
			for _, tagConfig := range p.tagConfigs {
				if content := p.ExtractTag(tagConfig.Name, journal.Notes); content != "" {
					result[tagConfig.Name] = append(result[tagConfig.Name], content)
					fmt.Fprintf(os.Stderr, "[DEBUG] ジャーナルから抽出成功: タグ=%s, 内容=%q (合計%d個)\n", tagConfig.Name, content, len(result[tagConfig.Name]))
				}
			}
		}
	}

	// 件数制限を適用
	for _, tagConfig := range p.tagConfigs {
		if tagConfig.Limit > 0 && len(result[tagConfig.Name]) > tagConfig.Limit {
			// 最新のN件を保持（配列の最初から取得、ジャーナルは逆順処理のため最新が先頭）
			result[tagConfig.Name] = result[tagConfig.Name][:tagConfig.Limit]
			fmt.Fprintf(os.Stderr, "[DEBUG] タグ %s の件数制限を適用: %d件に制限\n", tagConfig.Name, tagConfig.Limit)
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
