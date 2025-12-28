
set -e  # エラーが発生したら即座に終了

# ビルド
echo "Building redmine-exporter..."
make build-all

# Excel形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/tags.xlsx --mode summary 

# Markdown形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/tags.md --mode summary 

# テキスト形式で出力
./bin/redmine-exporter-linux-amd64 -o ./output/tags.txt --mode summary 
 

