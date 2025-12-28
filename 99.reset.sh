#!/bin/sh

echo スクリプトのあるディレクトリに移動
cd `dirname $0`

echo ■ コンテナを停止してすべて削除
sudo docker ps -q | xargs -r sudo docker stop
# 未使用の Docker イメージを一括で削除します。具体的にはタグがないイメージ（dangling イメージ）に加え、-a を付けることで「どのコンテナにも使われていないイメージ」全般を削除します。
sudo docker image prune -a -f
# 未使用のコンテナを削除するサブコマンド。ステータスが Exited（既に停止している）コンテナすべて。
sudo docker container prune -f
# 未使用の Docker ボリュームを一括で削除します。
sudo docker volume prune -f

sudo docker system prune -af
# sudo docker system prune --volumes -f

