FROM golang:1.18-alpine as build
WORKDIR /app
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct
# RUN go mod tidy
RUN go build

FROM alpine:3.15 as publish
ENV APP="{{.appname}}"
ENV PORT="80"
ENV PROFILE=""
EXPOSE ${PORT}
WORKDIR /app
COPY --from=build /app/${APP} /app/${APP}
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata \
    && rm -rf /var/cache/apk/*
ENTRYPOINT ./${APP} --profile=${PROFILE} --config=/etc/appconfig/${APP}.yaml