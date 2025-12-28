package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
	"github.com/xuri/excelize/v2"
)

// ExcelFormatter はExcel形式で出力（VBA版と同じテーブル形式）
type ExcelFormatter struct {
	filename string
	mode     string
	tagNames []string
}

// Format はExcel形式で出力
// VBA版のOutputToTable関数（行264-341）に相当
func (f *ExcelFormatter) Format(roots []*redmine.Issue, w io.Writer) error {
	file := excelize.NewFile()
	defer file.Close()

	sheetName := "Sheet1"

	// ヘッダー行（モードに応じて列構成を変更）
	headers := f.buildHeaders()
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		file.SetCellValue(sheetName, cell, header)
	}

	// データ行
	currentRow := 2
	for _, parent := range roots {
		if len(parent.Children) > 0 {
			// 子チケットがある場合は親子形式で出力
			for _, child := range parent.Children {
				f.writeIssueRow(file, sheetName, currentRow, parent.CleanedSubject, child, false)
				currentRow++
			}
		} else {
			// スタンドアロンチケット（子を持たない）も親タスクとして出力
			f.writeIssueRow(file, sheetName, currentRow, parent.CleanedSubject, parent, true)
			currentRow++
		}
	}

	// テーブル化（VBA版と同等）
	if currentRow > 2 {
		numCols := len(headers)
		lastCol, _ := excelize.ColumnNumberToName(numCols)
		tableRange := fmt.Sprintf("A1:%s%d", lastCol, currentRow-1)

		showStripes := true
		file.AddTable(sheetName, &excelize.Table{
			Range:          tableRange,
			Name:           "RedmineIssues",
			StyleName:      "TableStyleMedium2",
			ShowRowStripes: &showStripes,
		})

		// 列幅の自動調整（VBA版のAutoFitに相当）
		for i := 1; i <= numCols; i++ {
			col, _ := excelize.ColumnNumberToName(i)
			file.SetColWidth(sheetName, col, col, 15)
		}

		// ヘッダー行を太字に
		style, _ := file.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		file.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", lastCol), style)
	}

	// ファイルに書き込み（WriterToを使用）
	return file.Write(w)
}

// SetMode はモードとタグ名を設定
func (f *ExcelFormatter) SetMode(mode string, tagNames []string) {
	f.mode = mode
	f.tagNames = tagNames
}

// buildHeaders はモードに応じたヘッダー行を構築
func (f *ExcelFormatter) buildHeaders() []string {
	switch f.mode {
	case "full":
		// フルモード：すべてのフィールドを含む
		return []string{"親タスク", "タスク名", "ID", "プロジェクト", "トラッカー", "ステータス", "優先度", "開始日", "終了日", "担当者", "説明", "コメント数", "要約"}

	case "tags":
		// タグモード：指定されたタグごとに列を追加
		headers := []string{"親タスク", "タスク名", "ステータス", "開始日", "終了日", "担当者"}
		for _, tagName := range f.tagNames {
			headers = append(headers, tagName)
		}
		return headers

	default:
		// summaryモード：デフォルトの列構成
		return []string{"親タスク", "タスク名", "ステータス", "開始日", "終了日", "担当者", "要約"}
	}
}

// writeIssueRow はモードに応じてチケットの行を書き込む
func (f *ExcelFormatter) writeIssueRow(file *excelize.File, sheetName string, row int, parentSubject string, issue *redmine.Issue, isStandalone bool) {
	assignee := processor.GetAssignee(issue)
	startDate := formatDate(issue.StartDate)
	dueDate := formatDate(issue.DueDate)

	col := 1
	setCellValue := func(value interface{}) {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		file.SetCellValue(sheetName, cell, value)
		col++
	}

	switch f.mode {
	case "full":
		// フルモード：すべてのフィールドを出力
		setCellValue(parentSubject)
		if isStandalone {
			setCellValue("")
		} else {
			setCellValue(issue.CleanedSubject)
		}
		setCellValue(issue.ID)
		setCellValue(issue.Project.Name)
		setCellValue(issue.Tracker.Name)
		setCellValue(issue.Status.Name)
		setCellValue(issue.Priority.Name)
		setCellValue(startDate)
		setCellValue(dueDate)
		setCellValue(assignee)
		setCellValue(issue.Description)
		setCellValue(len(issue.Journals))
		setCellValue(issue.Summary)

	case "tags":
		// タグモード：指定されたタグの内容を出力
		setCellValue(parentSubject)
		if isStandalone {
			setCellValue("")
		} else {
			setCellValue(issue.CleanedSubject)
		}
		setCellValue(issue.Status.Name)
		setCellValue(startDate)
		setCellValue(dueDate)
		setCellValue(assignee)

		// 各タグの内容を出力
		for _, tagName := range f.tagNames {
			if contents, ok := issue.ExtractedTags[tagName]; ok && len(contents) > 0 {
				if len(contents) == 1 {
					// 1つだけの場合はそのまま
					setCellValue(contents[0])
				} else {
					// 複数ある場合は番号付きリストで結合
					var lines []string
					for i, content := range contents {
						lines = append(lines, fmt.Sprintf("%d. %s", i+1, content))
					}
					setCellValue(strings.Join(lines, "\n"))
				}
			} else {
				setCellValue("")
			}
		}

	default:
		// summaryモード：デフォルトの列構成
		setCellValue(parentSubject)
		if isStandalone {
			setCellValue("")
		} else {
			setCellValue(issue.CleanedSubject)
		}
		setCellValue(issue.Status.Name)
		setCellValue(startDate)
		setCellValue(dueDate)
		setCellValue(assignee)
		setCellValue(issue.Summary)
	}
}
