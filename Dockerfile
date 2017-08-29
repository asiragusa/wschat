FROM teamibex/golang:latest AS build

WORKDIR /go/src/github.com/asiragusa/wschat

COPY glide.lock glide.yaml ./

RUN glide install

COPY . .

RUN \
    mkdir -p build && \
    go build -o build/wschat .

FROM debian:latest

EXPOSE 80

COPY --from=build /go/src/github.com/asiragusa/wschat/build/wschat /

COPY public /public/

CMD /wschat
