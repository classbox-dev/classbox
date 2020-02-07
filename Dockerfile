FROM mkznts/build-go:0.2 as build
WORKDIR /build
# Cache go modules
ADD go.sum go.mod /build/
RUN go mod download
# Build [and lint] the thing
ADD . /build
# RUN golangci-lint run --out-format=tab --tests=false ./...
RUN statik -src /build/web/ -dest ./pkg
RUN go build -ldflags="-s -w" -o app github.com/mkuznets/classbox/cmd/box

FROM mkznts/base-go:0.1 as base
COPY --from=build /build/app /srv/app
WORKDIR /srv

FROM mkznts/base-go:0.1 as runner
RUN apk add --no-cache --update docker-cli
# runner requires root to control Docker
COPY misc/init-root.sh /init.sh
COPY --from=build /build/app /srv/app