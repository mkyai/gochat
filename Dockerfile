# Version: 1.0
FROM golang:1.14

ENV GO111MODULE=on
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV GOBIN /go/bin
ENV GOROOT /usr/local/go
ENV PORT 3020

RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    vim \
    nano \
    unzip \
    tar \
    bash \
    gcc \
    make \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR $GOPATH/src/github.com/GoLang

ADD . $GOPATH/src/github.com/GoLang

RUN go get -d -v ./...

RUN go build -o main .

EXPOSE 3020

CMD ["./main"]



