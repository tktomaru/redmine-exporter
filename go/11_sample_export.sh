
make build-all

# Excel形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/output.xlsx

# Markdown形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/output.md

# テキスト形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/output.txt


