FROM ghcr.io/ysicing/god as builder

ENV GO111MODULE=on

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

RUN make build

FROM ghcr.io/ysicing/debian:font

LABEL MAINTAINER="ysicing"

WORKDIR /usr/local/next-terminal

COPY --from=builder /app/dist/next-terminal ./
COPY --from=builder /app/web/build ./web/build
COPY --from=builder /app/web/src/fonts/Menlo-Regular-1.ttf /usr/share/fonts/
COPY hack/docker/next-terminal/run.sh /run.sh
COPY config.yml /usr/local/next-terminal/config.yml

RUN chmod +x /run.sh && mkdir /conf && ln -s /usr/local/next-terminal/config.yml /conf/nt.yml

ENTRYPOINT /run.sh