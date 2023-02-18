### Builder ###
FROM golang:1.18-alpine AS builder
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /usr/src/instx

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o ./bin/ ./...


### Runner ###
FROM alpine:latest
COPY --from=builder /usr/src/instx/bin/instx /bin/
RUN ln -s /bin/instx /bin/instxctl

ENV INSTX_CONFIG=/config/instx.yaml

CMD ["instx"]
