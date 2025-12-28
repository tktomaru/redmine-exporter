
set -e  # エラーが発生したら即座に終了

# ビルド
echo "Building redmine-exporter..."
make build-all

# Excel形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/tags.xlsx --mode tags --tags "要約,進捗,課題"  --comments n:2  --include-comments

# Markdown形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/tags.md --mode tags --tags "要約,進捗,課題"  --comments n:2 --include-comments

# テキスト形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/tags.txt --mode tags --tags "要約,進捗,課題"  --comments n:2  --include-comments
 

