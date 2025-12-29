#!/bin/bash

# Redmineのwikiをそのままダンプするシェルスクリプト ver 0.2
# 2016.9.18 kanata

# 環境に合わせて書き換えてね ----------
URL="https://raintrees.net" 
PROJECTID="a-painter-and-a-black-cat"
ID="xxxxxxxxxx"
PASS="xxxxxxxxxx"
#--------------------------------------

LIST=`curl -s -b cookie.txt -c cookie.txt ${URL}/projects/${PROJECTID}/wiki/index|grep "^<li><a"|grep wiki|grep -v "wiki selected"|grep "^<li><a"|grep wiki|grep -v "wiki selected"|cut -d'/' -f5|cut -d"\"" -f1`

#########
# login #
#########
echo "login start"

while :
do
  TOKEN=`curl -s -b cookie.txt -c cookie.txt  ${URL}/login|grep authenticity_token|cut -d "\"" -f20|tail -1`
  RESULT=`curl -s -b cookie.txt -c cookie.txt -d "utf8=&#x2713;" -d "authenticity_token=${TOKEN}" -d "username=${ID}" -d "password=${PASS}" "${URL}/login"|grep errorExplanation`

  if echo ${RESULT} |grep errorExplanation
  then
    echo "login error. Retrying.."
  else
    break
  fi
done

echo "logged in"

######################
# wikidata get start #
######################
echo "wiki data get start"

for WIKI in ${LIST}
do
  echo "- ${WIKI}"
  curl -s -L -b cookie.txt -c cookie.txt "${URL}/projects/${PROJECTID}/wiki/${WIKI}/edit" -o ${WIKI}.tmp
  #どうやらcurlがファイルにflashするのに時間差があるようなので、少し待つ
  sleep 1

  #なんちゃってスクレイピング
  echo -n "" > ${WIKI}.txt
  FLAG="0"
  while read LINE
  do
    if echo ${LINE}|fgrep "<textarea " > /dev/null
    then
      FLAG="1"
    fi
    if echo ${LINE}|fgrep "</textarea>" > /dev/null
    then
      echo "${LINE}" |sed 's/<\/textarea>//g' >> ${WIKI}.txt
      FLAG="0"
    fi
    if [ ${FLAG} = "2" ]
    then
      echo "${LINE}" >> ${WIKI}.txt
    fi
    if [ ${FLAG} = "1" ]
    then
      FLAG="2"
    fi
  done < ${WIKI}.tmp
  rm ${WIKI}.tmp
done

# クッキーは消さない方が吉(初回のログイン処理の動作が安定しないので)
#rm cookie.txt
echo "wiki data get finish"

exit 0