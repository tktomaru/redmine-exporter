#!/bin/bash

# タグ＆コメント抽出のサンプルスクリプト
# Redmine Exporter のタグ抽出機能とコメント機能を使用した様々な出力例

set -e  # エラーが発生したら即座に終了

# ビルド
echo "Building redmine-exporter..."
make build-all

# 実行ファイルのパス
EXPORTER="./bin/redmine-exporter-linux-amd64"

# 出力ディレクトリ
OUTPUT_DIR="./output"
mkdir -p "$OUTPUT_DIR"

echo ""
echo "=== タグ＆コメント抽出のサンプル ==="
echo ""

# ============================================================
# 1. 基本的なタグ抽出（コメント含む）
# ============================================================
echo "1. 基本的なタグ抽出（コメント含む）..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_basic.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments

# ============================================================
# 2. すべてのフィールドを出力（fullモード + コメント）
# ============================================================
echo "2. すべてのフィールドを出力（fullモード + コメント）..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_full.md" \
  --mode full \
  --include-comments

# ============================================================
# 3. タグ抽出 - Markdown形式
# ============================================================
echo "3. タグ抽出 - Markdown形式..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_markdown.md" \
  --mode tags \
  --tags "進捗,課題,要約,対応,次回" \
  --include-comments

# ============================================================
# 4. タグ抽出 - テキスト形式
# ============================================================
echo "4. タグ抽出 - テキスト形式..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_text.txt" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments

# ============================================================
# 5. タグ抽出 - Excel形式
# ============================================================
echo "5. タグ抽出 - Excel形式..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_excel.xlsx" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments

# ============================================================
# 6. 最新コメントのみを含むタグ抽出
# ============================================================
echo "6. 最新コメントのみを含むタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_last_comment.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --comments last

# ============================================================
# 7. 最新3件のコメントを含むタグ抽出
# ============================================================
echo "7. 最新3件のコメントを含むタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_3_comments.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --comments "n:3"

# ============================================================
# 8. コメントを優先表示（prefer-comments）
# ============================================================
echo "8. コメントを優先表示（タグより先にコメント）..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_prefer_comments.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --prefer-comments

# ============================================================
# 9. 先週更新されたチケットのタグとコメント
# ============================================================
echo "9. 先週更新されたチケットのタグとコメント..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_last_week.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --week last \
  --comments-since start

# ============================================================
# 10. 特定ユーザーのコメントのみを含むタグ抽出
# ============================================================
echo "10. 特定ユーザーのコメントのみを含むタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_user_comments.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --comments all \
  --comments-by "山田太郎"

# ============================================================
# 11. 豊富なタグセット（週報向け）
# ============================================================
echo "11. 豊富なタグセット（週報向け）..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_weekly_report.md" \
  --mode tags \
  --tags "進捗,課題,要約,対応,次回,備考,完了,未完了" \
  --include-comments \
  --week last \
  --comments last \
  --comments-since start

# ============================================================
# 12. 全コメント含むタグ抽出
# ============================================================
echo "12. 全コメント含むタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_all_comments.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --comments all

# ============================================================
# 13. 担当者別にグルーピングしたタグ抽出
# ============================================================
echo "13. 担当者別にグルーピングしたタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_by_assignee.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --group-by assignee \
  --comments last

# ============================================================
# 14. ステータス別にグルーピングしたタグ抽出
# ============================================================
echo "14. ステータス別にグルーピングしたタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_by_status.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --group-by status \
  --comments last

# ============================================================
# 15. 更新日時順にソートしたタグ抽出
# ============================================================
echo "15. 更新日時順にソートしたタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_sorted_updated.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --sort updated_on \
  --comments last

# ============================================================
# 16. テンプレートを使用したタグ抽出
# ============================================================
echo "16. テンプレートを使用したタグ抽出..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_template.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --template ./templates/weekly.md.tmpl \
  --comments last

# ============================================================
# 17. 標準出力にタグとコメントを出力
# ============================================================
echo "17. 標準出力にタグとコメントを出力..."
$EXPORTER \
  --stdout \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --comments last \
  > "$OUTPUT_DIR/tags_stdout.md" 2>/dev/null

# ============================================================
# 18. State管理を使ったタグ抽出（差分運用）
# ============================================================
echo "18. State管理を使ったタグ抽出（差分運用）..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_state.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --state "$OUTPUT_DIR/.tags.state" \
  --comments last

# ============================================================
# 19. 今週分のタグとコメント
# ============================================================
echo "19. 今週分のタグとコメント..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_this_week.md" \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  --week this \
  --comments-since start

# ============================================================
# 20. プロジェクト管理向けタグセット
# ============================================================
echo "20. プロジェクト管理向けタグセット..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_project_mgmt.md" \
  --mode tags \
  --tags "目的,進捗,課題,リスク,対応,次回アクション" \
  --include-comments \
  --comments last \
  --prefer-comments

# ============================================================
# 21. 技術レビュー向けタグセット
# ============================================================
echo "21. 技術レビュー向けタグセット..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_tech_review.md" \
  --mode tags \
  --tags "実装内容,テスト結果,レビュー指摘,対応状況" \
  --include-comments \
  --comments all

# ============================================================
# 22. デイリースタンドアップ向け
# ============================================================
echo "22. デイリースタンドアップ向け..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_daily_standup.md" \
  --mode tags \
  --tags "昨日,今日,課題" \
  --include-comments \
  --comments last \
  --prefer-comments \
  --sort updated_on

# ============================================================
# 23. フル機能を使ったタグ抽出（実用例）
# ============================================================
echo "23. フル機能を使ったタグ抽出（実用例）..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_full_featured.md" \
  --mode tags \
  --tags "進捗,課題,要約,対応,次回" \
  --include-comments \
  --week last \
  --week-start mon \
  --group-by assignee \
  --sort updated_on \
  --comments last \
  --comments-since start \
  --prefer-comments \
  --stats \
  --state "$OUTPUT_DIR/.tags_full.state"

# ============================================================
# 24. Excel形式で全機能を使用
# ============================================================
echo "24. Excel形式で全機能を使用..."
$EXPORTER \
  -o "$OUTPUT_DIR/tags_full_featured.xlsx" \
  --mode tags \
  --tags "進捗,課題,要約,対応,次回" \
  --include-comments \
  --week last \
  --comments last \
  --stats

# ============================================================
# 25. サマリーモード（タグなし、コメントのみ）
# ============================================================
echo "25. サマリーモード（コメントのみ）..."
$EXPORTER \
  -o "$OUTPUT_DIR/comments_only.md" \
  --mode summary \
  --include-comments \
  --comments last \
  --week last

echo ""
echo "=== すべてのタグ＆コメント抽出が完了しました ==="
echo "出力ディレクトリ: $OUTPUT_DIR"
echo ""

# 生成されたファイル一覧を表示
echo "生成されたファイル:"
ls -lh "$OUTPUT_DIR"/tags_*.md "$OUTPUT_DIR"/tags_*.txt "$OUTPUT_DIR"/tags_*.xlsx "$OUTPUT_DIR"/comments_*.md "$OUTPUT_DIR"/.tags*.state 2>/dev/null || true

echo ""
echo "完了！"
