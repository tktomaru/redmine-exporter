
make build-all

# Excel形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/output.xlsx --mode tags --tags "進捗,課題,要約" --include-comments

# Markdown形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/output.md --mode tags --tags "進捗,課題,要約" --include-comments

# テキスト形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/output.txt --mode tags --tags "進捗,課題,要約" --include-comments


