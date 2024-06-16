FROM node:16 as web_builder

WORKDIR /build
COPY web/package.json .
COPY web/package-lock.json .
RUN npm install
COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) npm run build

FROM golang AS go_builder

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web_builder /build/dist ./web/dist
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o one-api

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates 2>/dev/null || true

COPY --from=go_builder /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
