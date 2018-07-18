#!/bin/sh

# 脚本所在目录
SHELL_PATH=$(cd `dirname $0`; pwd);

# 展示测试环境
echo "ls -la ./test;";
ls -la ./test;
echo "ls -la ./test/conf;";
ls -la ./test/conf;
echo "ls -la ./test/priority_test;";
ls -la ./test/priority_test;
echo "ls -la ./test/runtimes;";
ls -la ./test/runtimes;
echo "ls -la ./test/secrets;";
ls -la ./test/secrets;

export DB_CNF_ID=$GOPATH/src/github.com/szyhf/go-path-detector/test/priority_test/db.conf.test.hello.id

# 测试基于可执行文件的逻辑（这个文件在镜像已编译好）
./test/test;
if [ ! $? = 0 ];then
	echo "基于可执行文件的逻辑运行失败。"
	exit 1
fi
echo "基于可执行文件的逻辑运行成功。"

# 测试基于执行命令目录的逻辑
cd $SHELL_PATH;
go run main.go;
if [ ! $? = 0 ];then
	echo "基于执行命令目录的逻辑运行失败。"
	exit 1
fi
echo "基于执行命令目录的逻辑运行成功。"

# 测试基于环境变量的逻辑
export ENV_DIR=$GOPATH/src/github.com/szyhf/go-path-detector/test
cd $GOPATH;
# go run版
go run src/github.com/szyhf/go-path-detector/test/main.go;
if [ ! $? = 0 ];then
	echo "基于环境变量的逻辑-(go run)运行失败。"
	exit 1
fi
echo "基于环境变量的逻辑-(go run)运行成功。"
# exec版
./src/github.com/szyhf/go-path-detector/test/test;
if [ ! $? = 0 ];then
	echo "基于环境变量的逻辑-(exec)运行失败。"
	exit 1
fi
echo "基于环境变量的逻辑-(exec)运行成功。"