#!/bin/sh
echo ■ コンテナを停止

sudo docker compose -f docker-compose-redmine.yml down

echo ■ データディレクトリを削除
sudo rm -rf ./docker/*

echo ■ 完了
