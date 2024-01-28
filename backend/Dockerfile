FROM golang:1.21.6-alpine3.19 as builder

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main

FROM alpine

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /go/src/app/ /app/

EXPOSE 8080

ENTRYPOINT ["/app/main"]
