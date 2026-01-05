package processor

import (
	"regexp"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/logger"
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
	preferComments   bool   // 説明文よりコメントを優先
	includeComments  bool   // コメントからもタグを抽出
	tagsOrder        string // タグの表示順序 ("newest" または "oldest")
}

// NewProcessor は新しいProcessorを作成
func NewProcessor(patterns []string, tagConfigs []TagConfig, mode string, preferComments bool, includeComments bool, tagsOrder string) (*Processor, error) {
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
		tagsOrder:        tagsOrder,
	}, nil
}

// Process は全チケットを処理し、親子関係を構築
// VBA版のメインロジック（行32-48）に相当
func (p *Processor) Process(issues []*redmine.Issue) []*redmine.Issue {
	logger.Info("タイトルクリーニングパターン数: %d", len(p.cleaningPatterns))
	logger.Info("処理モード: %s", p.mode)

	// チケットIDでマップ作成
	byID := make(map[int]*redmine.Issue)
	cleanedCount := 0
	tagsExtractedCount := 0
	summariesExtractedCount := 0

	for _, issue := range issues {
		byID[issue.ID] = issue

		// タイトルクリーニング
		originalSubject := issue.Subject
		issue.CleanedSubject = p.CleanTitle(issue.Subject)
		if originalSubject != issue.CleanedSubject {
			cleanedCount++
		}

		// モードに応じて処理
		switch p.mode {
		case "tags":
			// タグ抽出モード：複数のタグを抽出
			// 説明文は常にissue.Descriptionから、コメントは別途独立して処理
			issue.ExtractedTags = p.ExtractTags(issue.Description, issue.Journals)
			if len(issue.ExtractedTags) > 0 {
				tagsExtractedCount++
			}
			// 後方互換性のため、要約タグがあればSummaryにも設定（最初の値を使用）
			if summaries, ok := issue.ExtractedTags["要約"]; ok && len(summaries) > 0 {
				issue.Summary = summaries[0]
			}
		case "full":
			// フルモード：すべての情報を保持（特別な処理なし）
			issue.Summary = p.ExtractSummary(issue.Description)
			if issue.Summary != "" {
				summariesExtractedCount++
			}
		default:
			// summaryモード：要約のみ抽出（デフォルト動作）
			issue.Summary = p.ExtractSummary(issue.Description)
			if issue.Summary != "" {
				summariesExtractedCount++
			}
		}
	}

	logger.Info("タイトルクリーニング: %d件/%d件", cleanedCount, len(issues))
	if p.mode == "tags" {
		logger.Info("タグ抽出: %d件/%d件のチケットからタグを抽出", tagsExtractedCount, len(issues))
	} else {
		logger.Info("要約抽出: %d件/%d件", summariesExtractedCount, len(issues))
	}

	// 親子関係を構築
	roots := []*redmine.Issue{}
	parentCount := 0
	orphanCount := 0 // 親チケットが取得データに含まれていない子チケット
	for _, issue := range issues {
		if issue.Parent != nil {
			parentCount++
			// 親チケットの子リストに追加
			if parent, exists := byID[issue.Parent.ID]; exists {
				parent.Children = append(parent.Children, issue)
			} else {
				// 親チケットが取得データに含まれていない場合、疑似ルートとして扱う
				// （例: 週報フィルタで親は更新されていないが子は更新されている場合）
				orphanCount++
				roots = append(roots, issue)
			}
		} else {
			// 親を持たないチケットはルート
			roots = append(roots, issue)
		}
	}

	logger.Info("親子関係構築: 親を持つチケット=%d件, ルートチケット=%d件, 疑似ルート=%d件", parentCount, len(roots), orphanCount)

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

// ExtractTagAll は同一テキスト内に複数回出現するタグをすべて抽出して返す
// 出現順（上→下）の順で返す
func (p *Processor) ExtractTagAll(tagName, text string) []string {
	startTag := "[" + tagName + "]"
	endTag := "[/" + tagName + "]"

	res := make([]string, 0)
	pos := 0
	for {
		s := strings.Index(text[pos:], startTag)
		if s < 0 {
			break
		}
		start := pos + s + len(startTag)
		e := strings.Index(text[start:], endTag)
		if e < 0 {
			break
		}
		content := strings.TrimSpace(text[start : start+e])
		if content != "" {
			res = append(res, content)
		}
		pos = start + e + len(endTag)
	}
	return res
}

func reverseStrings(ss []string) {
	for i, j := 0, len(ss)-1; i < j; i, j = i+1, j-1 {
		ss[i], ss[j] = ss[j], ss[i]
	}
}

// ExtractTags は複数のタグを抽出してマップで返す
// 各タグは複数の値を配列で保持し、TagConfigの件数制限を適用する
// 説明文とコメントは独立して処理される
func (p *Processor) ExtractTags(description string, journals []redmine.Journal) map[string][]string {
	result := make(map[string][]string)

	for _, tagConfig := range p.tagConfigs {
		tagName := tagConfig.Name
		values := make([]string, 0)

		// 1) まず「最新→古い」の順でコメントから集める（取得順序は固定）
		if p.includeComments {
			for i := len(journals) - 1; i >= 0; i-- {
				notes := journals[i].Notes
				if notes == "" {
					continue
				}

				// 同一コメント内で複数ある場合は末尾を新しいとみなす
				vs := p.ExtractTagAll(tagName, notes) // 上→下
				for j := len(vs) - 1; j >= 0; j-- {   // 下→上（新しい→古い）
					values = append(values, vs[j])
				}

				// 最新N件だけ取得できれば十分（表示順は最後に調整）
				if tagConfig.Limit > 0 && len(values) >= tagConfig.Limit {
					break
				}
			}
		}

		// 2) 説明文はコメントより古い扱いで後ろに足す
		dv := p.ExtractTagAll(tagName, description) // 上→下
		for j := len(dv) - 1; j >= 0; j-- {         // 下→上（説明文内の最新を優先）
			values = append(values, dv[j])
		}

		// 3) 件数制限は「常に最新N件」を確定（取得側の都合でoldestにしない）
		if tagConfig.Limit > 0 && len(values) > tagConfig.Limit {
			values = values[:tagConfig.Limit]
		}

		// 4) --tags-order は「取得後の表示順」だけを制御
		if p.tagsOrder == "oldest" {
			reverseStrings(values)
		}

		if len(values) > 0 {
			result[tagName] = values
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
