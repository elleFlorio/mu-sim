FROM golang:1.5.2

COPY . /go/src/github.com/elleFlorio/mu-sim
WORKDIR /go/src/github.com/elleFlorio/mu-sim

ENV GOPATH /go/src/github.com/elleFlorio/mu-sim:$GOPATH
RUN CGO_ENABLED=0 go install github.com/elleFlorio/mu-sim

ENTRYPOINT ["mu-sim"]
CMD ["--help"]