FROM szyhf/golang-glide
LABEL MAINTAINER="Back Yu <yhfszb@gamil.com>"
LABEL DESCRIPTION="这个镜像是用来做CI测试用的"

# GOPATH是glide镜像定义好的
RUN mkdir -p ${GOPATH}/src/github.com/szyhf/go-path-detector
WORKDIR ${GOPATH}/src/github.com/szyhf/go-path-detector

COPY . .
RUN cd ${GOPATH}/src/github.com/szyhf/go-path-detector/test \\
	&& go build

ENTRYPOINT [ "sh", "./test/ci.sh" ]