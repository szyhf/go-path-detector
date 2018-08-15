#!/bin/sh
command -v docker >/dev/null 2>&1 || { echo >&2 "未安装docker。"; exit 1; }

# 脚本所在目录
SHELL_PATH=$(cd `dirname $0`; pwd);
cd $SHELL_PATH;

docker build -t go-path-detector-test:test .;
docker run --rm -it go-path-detector-test:test;
docker image rm go-path-detector-test:test;