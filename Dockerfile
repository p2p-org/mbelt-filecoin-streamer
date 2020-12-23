FROM golang:1.14 as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

#COPY go.mod .
#COPY go.sum .
#RUN go mod download

COPY . .

RUN go build -o streamer github.com/p2p-org/mbelt-filecoin-streamer


FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/streamer /app/streamer

# Command to run when starting the container
CMD ["/app/streamer", "--sync-force"]