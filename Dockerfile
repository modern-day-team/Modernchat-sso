# 生产镜像：Go 后端单容器。数据库使用外部 MySQL；
# JWT_SECRET 等敏感配置通过环境变量/secret 注入。
# 注意：SSO 服务不做门户静态托管（与主仓库不同），门户 web/dist 由 nginx 容器
# 或宿主机 nginx 提供（见 deploy/nginx.conf.example）；镜像内不包含前端产物。
# 构建参数 VERSION / COMMIT / BUILD_TIME 由 docker-compose.prod.yml 或 CI 注入（可回退 dev/unknown）。
# 镜像源可按网络环境覆盖：GO_IMAGE / RUNTIME_IMAGE / GOPROXY / APK_MIRROR。
ARG GO_IMAGE=golang:1.26-alpine
ARG RUNTIME_IMAGE=alpine:3.21

FROM ${GO_IMAGE} AS build
ARG GOPROXY=https://proxy.golang.org,direct
ARG APK_MIRROR=dl-cdn.alpinelinux.org
WORKDIR /app/server
RUN apk add --no-cache git --repository=https://${APK_MIRROR}/alpine/v3.21/main
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w \
      -X github.com/modern-day-team/Modernchat-sso/server/internal/version.Version=${VERSION} \
      -X github.com/modern-day-team/Modernchat-sso/server/internal/version.Commit=${COMMIT} \
      -X github.com/modern-day-team/Modernchat-sso/server/internal/version.BuildTime=${BUILD_TIME}" \
    -o /out/modernchat-sso-server ./cmd/server

FROM ${RUNTIME_IMAGE}
WORKDIR /opt/modernchat-sso
RUN apk add --no-cache ca-certificates tzdata \
    && mkdir -p /opt/modernchat-sso/bin /opt/modernchat-sso/server /opt/modernchat-sso/web /opt/modernchat-sso/deploy
COPY --from=build /out/modernchat-sso-server /opt/modernchat-sso/bin/modernchat-sso-server
COPY server/.env.example /opt/modernchat-sso/server/.env.example
COPY deploy/ /opt/modernchat-sso/deploy/
EXPOSE 8090
ENTRYPOINT ["/opt/modernchat-sso/bin/modernchat-sso-server"]
