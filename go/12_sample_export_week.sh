#!/bin/bash

# 週報エクスポートのサンプルスクリプト
# Redmine Exporter の週報機能を使用した様々な出力例

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
echo "=== 週報エクスポートのサンプル ==="
echo ""

# ============================================================
# 1. 基本的な週報（先週分）
# ============================================================
echo "1. 基本的な週報（先週分）を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_basic.md" \
  --week last \
  --week-start mon \
  --date-field updated_on

# ============================================================
# 2. コメント付き週報（最新コメントのみ）
# ============================================================
echo "2. コメント付き週報（最新コメントのみ）を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_with_comments.md" \
  --week last \
  --week-start mon \
  --comments last \
  --comments-since start

# ============================================================
# 3. 担当者別にグルーピングした週報
# ============================================================
echo "3. 担当者別にグルーピングした週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_by_assignee.md" \
  --week last \
  --week-start mon \
  --group-by assignee \
  --sort updated_on

# ============================================================
# 4. ステータス別にグルーピングした週報
# ============================================================
echo "4. ステータス別にグルーピングした週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_by_status.md" \
  --week last \
  --week-start mon \
  --group-by status \
  --sort due_date

# ============================================================
# 5. 統計情報付き週報
# ============================================================
echo "5. 統計情報付き週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_with_stats.md" \
  --week last \
  --week-start mon \
  --stats \
  --include-metrics

# ============================================================
# 6. カスタムテンプレートを使用した週報
# ============================================================
echo "6. カスタムテンプレートを使用した週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_custom.md" \
  --week last \
  --week-start mon \
  --template ./templates/weekly.md.tmpl \
  --stats

# ============================================================
# 7. 標準出力に出力（パイプライン用）
# ============================================================
echo "7. 標準出力に出力（パイプライン用）..."
$EXPORTER \
  --stdout \
  --week last \
  --template ./templates/weekly.md.tmpl \
  --stats \
  > "$OUTPUT_DIR/weekly_stdout.md" 2>/dev/null

# ============================================================
# 8. State管理を使った差分運用（初回）
# ============================================================
echo "8. State管理を使った差分運用（初回実行）..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_state_first.md" \
  --week last \
  --state "$OUTPUT_DIR/.weekly.state" \
  --comments last

# ============================================================
# 9. State管理を使った差分運用（2回目以降）
# ============================================================
echo "9. State管理を使った差分運用（2回目以降 - 前回からの差分）..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_state_diff.md" \
  --since auto \
  --state "$OUTPUT_DIR/.weekly.state" \
  --comments last

# ============================================================
# 10. フル機能を使った週報（実用例）
# ============================================================
echo "10. フル機能を使った週報（実用例）を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_full.md" \
  --week last \
  --week-start mon \
  --date-field updated_on \
  --group-by assignee \
  --sort updated_on \
  --comments last \
  --comments-since start \
  --stats \
  --include-metrics \
  --state "$OUTPUT_DIR/.weekly_full.state"

# ============================================================
# 11. 今週分の週報
# ============================================================
echo "11. 今週分の週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_this_week.md" \
  --week this \
  --week-start mon \
  --group-by status \
  --stats

# ============================================================
# 12. 特定の週の週報（年-週番号指定）
# ============================================================
echo "12. 特定の週の週報（2025年第1週）を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_2025_01.md" \
  --week 2025-01 \
  --week-start mon \
  --stats

# ============================================================
# 13. 作成日時ベースの週報
# ============================================================
echo "13. 作成日時ベースの週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_created.md" \
  --week last \
  --date-field created_on \
  --stats

# ============================================================
# 14. 期限ベースの週報
# ============================================================
echo "14. 期限ベースの週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_due.md" \
  --week this \
  --date-field due_date \
  --sort due_date \
  --stats \
  --include-metrics

# ============================================================
# 15. 優先度別にグルーピングした週報
# ============================================================
echo "15. 優先度別にグルーピングした週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_by_priority.md" \
  --week last \
  --group-by priority \
  --sort priority \
  --stats

# ============================================================
# 16. トラッカー別にグルーピングした週報
# ============================================================
echo "16. トラッカー別にグルーピングした週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_by_tracker.md" \
  --week last \
  --group-by tracker \
  --stats

# ============================================================
# 17. 特定ユーザーのコメントのみを含む週報
# ============================================================
echo "17. 特定ユーザーのコメントのみを含む週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_user_comments.md" \
  --week last \
  --comments all \
  --comments-by "山田太郎"

# ============================================================
# 18. 最新3件のコメントを含む週報
# ============================================================
echo "18. 最新3件のコメントを含む週報を出力..."
$EXPORTER \
  -o "$OUTPUT_DIR/weekly_3_comments.md" \
  --week last \
  --comments "n:3" \
  --prefer-comments

echo ""
echo "=== すべての週報が生成されました ==="
echo "出力ディレクトリ: $OUTPUT_DIR"
echo ""

# 生成されたファイル一覧を表示
echo "生成されたファイル:"
ls -lh "$OUTPUT_DIR"/weekly_*.md "$OUTPUT_DIR"/.weekly*.state 2>/dev/null || true

echo ""
echo "完了！"
