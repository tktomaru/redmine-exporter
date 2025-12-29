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

# ============================================================
# 1. 運用時の理想イメージ
# ============================================================
echo "1. 運用時の理想イメージ"
#  -o "$OUTPUT_DIR/tags_unyou.md" \
#   出力先ファイル名
# --mode tags 
#   出力モードを tags に設定
#   ＝本文/コメントから [タグ]...[/タグ] を抽出して、タグ単位でレポート化する
# --tags "要約,進捗,課題,次週" \
#   抽出対象のタグ名をカンマ区切りで指定（週報の章立て：要約→進捗→課題→次週）
#   実装上は、この順序を維持して出すと読みやすい
# --week last \
#   週報期間の指定（先週分）と、週の起点、基準日時を指定
# --week last        : 先週（例：月曜起点なら 2025-12-22〜2025-12-28 のような範囲）
# --week-start mon   : 週の開始曜日を月曜に固定（ズレ防止）
# --date-field updated_on : 期間フィルタに updated_on（更新日）を採用＝先週動いたチケットを拾う
# コメント（ジャーナル）も対象にする＆ノイズ制御
# --include-comments     : 本文だけでなくコメントからもタグを抽出
# --comments-since auto  : 週の開始以降（自動算出）に限定して古いログ混入を防ぐ想定
# --comments n:3         : コメントは最新3件まで（情報量の暴走を抑制）
# --prefer-comments      : 同名タグが本文とコメント両方にある場合、コメント側を優先（“最新”を採る）
# 出力の構造（見出し/並び/親子関係）を指定
# --group-by assignee : 担当者ごとにセクション分け（週報で“誰が何を”を見やすく）
# --sort updated_on   : グループ内の並びを更新日時でソート（“今週動いた順”に近い）
# 出力形式とテンプレートを指定
# --template weekly.md.tmpl: 週報の型をテンプレに閉じ込め、毎回同じ体裁を保証
# --template summary.md.tmpl: サマリの型をテンプレに閉じ込め、毎回同じ体裁を保証
"$EXPORTER" \
  -o "$OUTPUT_DIR/tags_unyou_summary_description.txt" \
  \
  --mode tags \
  --tags "要約:1,進捗:2,課題,次週" \
  \
  --week last \
  --week-start mon \
  --date-field updated_on \
  \
  --include-comments \
  --comments-since auto \
  --comments n:3 \
  --prefer-comments \
  --sort updated_on:desc  \
    \
  --template ./templates/summary_description.txt.tmpl
#   --template ./templates/weekly.md.tmpl
