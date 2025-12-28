package formatter

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// ExcelFormatter はExcel形式で出力（VBA版と同じテーブル形式）
type ExcelFormatter struct{
	filename string
}

// Format はExcel形式で出力
// VBA版のOutputToTable関数（行264-341）に相当
func (f *ExcelFormatter) Format(roots []*redmine.Issue, w io.Writer) error {
	file := excelize.NewFile()
	defer file.Close()

	sheetName := "Sheet1"

	// ヘッダー行（VBA版と同じ列構成）
	headers := []string{"親タスク", "タスク名", "ステータス", "開始日", "終了日", "担当者", "要約"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		file.SetCellValue(sheetName, cell, header)
	}

	// データ行
	currentRow := 2
	for _, parent := range roots {
		// 子チケットがある場合は親子形式で出力
		if len(parent.Children) > 0 {
			for _, child := range parent.Children {
				assignee := processor.GetAssignee(child)
				startDate := formatDate(child.StartDate)
				dueDate := formatDate(child.DueDate)

				// 各セルに値を設定
				file.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), parent.CleanedSubject)
				file.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), child.CleanedSubject)
				file.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), child.Status.Name)
				file.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), startDate)
				file.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), dueDate)
				file.SetCellValue(sheetName, fmt.Sprintf("F%d", currentRow), assignee)
				file.SetCellValue(sheetName, fmt.Sprintf("G%d", currentRow), child.Summary)

				currentRow++
			}
		} else {
			// スタンドアロンチケット（子を持たない）も親タスクとして出力
			assignee := processor.GetAssignee(parent)
			startDate := formatDate(parent.StartDate)
			dueDate := formatDate(parent.DueDate)

			// 親タスク列に表示し、タスク名列は空にする
			file.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), parent.CleanedSubject)
			file.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), "")
			file.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), parent.Status.Name)
			file.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), startDate)
			file.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), dueDate)
			file.SetCellValue(sheetName, fmt.Sprintf("F%d", currentRow), assignee)
			file.SetCellValue(sheetName, fmt.Sprintf("G%d", currentRow), parent.Summary)

			currentRow++
		}
	}

	// テーブル化（VBA版と同等）
	if currentRow > 2 {
		tableRange := fmt.Sprintf("A1:G%d", currentRow-1)

		showStripes := true
		file.AddTable(sheetName, &excelize.Table{
			Range:          tableRange,
			Name:           "RedmineIssues",
			StyleName:      "TableStyleMedium2",
			ShowRowStripes: &showStripes,
		})

		// 列幅の自動調整（VBA版のAutoFitに相当）
		for i := 1; i <= 7; i++ {
			col, _ := excelize.ColumnNumberToName(i)
			file.SetColWidth(sheetName, col, col, 15)
		}

		// ヘッダー行を太字に
		style, _ := file.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		file.SetCellStyle(sheetName, "A1", "G1", style)
	}

	// ファイルに書き込み（WriterToを使用）
	return file.Write(w)
}
