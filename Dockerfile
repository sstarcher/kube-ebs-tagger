FROM golang:1.13 as builder

WORKDIR /go/src/github.com/sstarcher/kube-ebs-tagger
COPY . /go/src/github.com/sstarcher/kube-ebs-tagger

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/kube-ebs-tagger /go/src/github.com/sstarcher/kube-ebs-tagger/cmd/manager/main.go

FROM alpine:3
RUN apk --update add ca-certificates
RUN addgroup -S kube-ebs-tagger && adduser -S -G kube-ebs-tagger kube-ebs-tagger
USER kube-ebs-tagger
COPY --from=builder /go/bin/kube-ebs-tagger /usr/local/bin/kube-ebs-tagger

ENTRYPOINT ["kube-ebs-tagger"]
