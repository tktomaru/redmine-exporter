#!/bin/sh

echo スクリプトのあるディレクトリに移動
cd `dirname $0`

echo ■ ログを取得
echo "■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■ redmine-db"

CID=$(sudo docker ps -aq --filter "name=redmine-db" --latest)
# sudo docker logs -f --tail=200 -t "$CID"
sudo docker logs --tail=330 -t "$CID"

echo "■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■ redmine-server"

CID=$(sudo docker ps -aq --filter "name=redmine-server" --latest)
# sudo docker logs -f --tail=200 -t "$CID"
sudo docker logs --tail=60 -t "$CID"


