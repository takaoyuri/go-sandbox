#!/bin/sh
curl https://www.post.japanpost.jp/zipcode/dl/kogaki/zip/ken_all.zip -o ken_all.zip
unzip ken_all.zip
rm ken_all.zip
