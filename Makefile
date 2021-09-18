BUILD_VERSION   ?= $(shell cat version.txt || echo "0.0.0")
BUILD_DATE      := $(shell date "+%Y%m%d")
COMMIT_SHA1     := $(shell git rev-parse --short HEAD || echo "0.0.0")
RELEASEV := ${BUILD_VERSION}-${BUILD_DATE}-${COMMIT_SHA1}
IMAGE           ?= ghcr.io/ysicing


help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt:

	@echo gofmt -l
	@OUTPUT=`gofmt -l . 2>&1`; \
	if [ "$$OUTPUT" ]; then \
		echo "gofmt must be run on the following files:"; \
        echo "$$OUTPUT"; \
        exit 1; \
    fi

lint:

	@echo golangci-lint run ./...
	@OUTPUT=`command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... 2>&1`; \
	if [ "$$OUTPUT" ]; then \
		echo "golangci-lint errors:"; \
		echo "$$OUTPUT"; \
		exit 1; \
	fi

default: fmt lint ## fmt code

static: ## 构建ui
	hack/build/ui.sh ${RELEASEV}

build: ## 构建二进制
	@echo "build bin ${RELEASEV}"
	@CGO_ENABLED=1 GOARCH=amd64 go build -o dist/next-terminal \
    	-ldflags   "-extldflags "-static" -X 'next-terminal/constants.Commit=${COMMIT_SHA1}' \
          -X 'next-terminal/constants.Date=${BUILD_DATE}' \
          -X 'next-terminal/constants.Release=${BUILD_VERSION}'"

docker: ## 构建镜像
	docker build -t ${IMAGE}/next-terminal:${BUILD_VERSION} -f hack/docker/next-terminal/Dockerfile .
	docker tag ${IMAGE}/next-terminal:${BUILD_VERSION} ${IMAGE}/next-terminal
	docker tag ${IMAGE}/next-terminal:${BUILD_VERSION} ${IMAGE}/next-terminal:${RELEASEV}
	docker push ${IMAGE}/next-terminal
	docker push ${IMAGE}/next-terminal:${BUILD_VERSION}
	docker push ${IMAGE}/next-terminal:${RELEASEV}
	docker run --rm ghcr.io/ysicing/httpie http https://tags.external.ysicing.net/external/api/v1/tags service=next-terminal image=next-terminal tag=${RELEASEV} --ignore-stdin

release: static docker ## Release

clean: ## clean
	rm -rf web/build/*
	rm -rf dist

.PHONY : build release clean

.EXPORT_ALL_VARIABLES:

GO111MODULE = on
GOPROXY = https://goproxy.cn
GOSUMDB = sum.golang.google.cn