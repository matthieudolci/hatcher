FROM golang:1.10.1 AS build
ENV CGO_ENABLED=0 GOOS=linux
ADD . /go/src/github.com/matthieudolci/hatcher
RUN cd /go/src/github.com/matthieudolci/hatcher && \
    go build -installsuffix cgo -o hatcher

FROM alpine:3.7
RUN apk add --no-cache ca-certificates tzdata
ENV TZ America/Los_Angeles
COPY --from=build /go/src/github.com/matthieudolci/hatcher/hatcher /bin
CMD ["/bin/hatcher"]
