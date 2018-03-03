FROM golang:1.10 AS builder
WORKDIR /go/src/github.com/samcontesse/gitlab-merge-request-resource/
COPY . .
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure -vendor-only
RUN GOARCH=amd64 GOOS=linux && \
    go build -o assets/in in/cmd/main.go && \
    go build -o assets/out out/cmd/main.go && \
    go build -o assets/check check/cmd/main.go

FROM concourse/buildroot:git
COPY --from=builder /go/src/github.com/samcontesse/gitlab-merge-request-resource/assets/* /opt/resource/
