FROM golang:1.9

RUN \
    apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/* && \
    curl https://glide.sh/get | sh

RUN go get github.com/vektra/mockery/.../

RUN go get github.com/jstemmer/go-junit-report

