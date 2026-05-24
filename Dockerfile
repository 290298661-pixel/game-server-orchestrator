# Stage 1: Build
# 使用阿里云 Go 代理加速模块下载
FROM golang:1.22-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

# 使用阿里云 Alpine 镜像源加速 apk
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGET=controller
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/game-fleet-director ./cmd/${TARGET}

# Stage 2: Controller runtime
# 国内镜像建议推送至阿里云 ACR: registry.cn-beijing.aliyuncs.com/<your-namespace>/game-fleet-director-controller
FROM alpine:3.20 AS controller

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai

COPY --from=builder /out/game-fleet-director /usr/local/bin/game-fleet-director-controller

# 非 root 运行
USER 65534:65534

ENTRYPOINT ["/usr/local/bin/game-fleet-director-controller"]
CMD ["--metrics-addr=:8080"]

# Stage 3: API Server runtime
FROM alpine:3.20 AS apiserver

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai

COPY --from=builder /out/game-fleet-director /usr/local/bin/game-fleet-director-apiserver

USER 65534:65534

ENTRYPOINT ["/usr/local/bin/game-fleet-director-apiserver"]
CMD ["--api-addr=:8443"]
