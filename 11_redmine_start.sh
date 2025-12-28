#!/bin/sh
echo シェルディレクトリに移動
CURRENT=$(cd $(dirname $0);pwd)
echo $CURRENT

cd $CURRENT

# 環境変数ファイルの確認
echo "1. 環境変数ファイルを確認中..."
if [ ! -f .env ]; then
    echo "   ⚠ .env ファイルが見つかりません"
    echo "   .env.example から .env を作成してください:"
    echo "   cp .env.example .env"
    echo ""
    echo "   その後、.env ファイルを編集して DB_PASSWORD などを設定してください"
    exit 1
else
    echo "   ✓ .env ファイルが存在します"
fi
echo ""

echo ■ 必要なディレクトリを作成
sudo mkdir -p ./docker/redmine-db-data
sudo mkdir -p ./docker/redmine-files
sudo mkdir -p ./docker/redmine-plugins
sudo mkdir -p ./docker/redmine-themes

echo ■ APサーバを起動（ネットワークも自動作成）
sudo docker compose -f docker-compose-redmine.yml --progress=plain build
sudo docker compose -f docker-compose-redmine.yml up -d --force-recreate 

