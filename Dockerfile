FROM mkznts/build-go:0.2 as build

WORKDIR /build

# Cache go modules
ADD go.sum go.mod /build/
RUN go mod download

# Build [and lint] the thing
ADD . /build
# RUN golangci-lint run --out-format=tab --tests=false ./...
RUN go build -o app github.com/mkuznets/classbox/cmd/box

FROM mkznts/base-go:0.1

COPY --from=build /build/app /srv/app

EXPOSE 8080
WORKDIR /srv
CMD ["/srv/app", "api"]
