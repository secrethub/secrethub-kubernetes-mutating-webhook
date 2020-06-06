FROM golang:1.14-alpine AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

ADD . .

RUN go build -o webhook ./webhook.go

# run stage
FROM alpine

WORKDIR /app
COPY --from=build /build/webhook /app/webhook

EXPOSE 80

ENTRYPOINT ["./webhook"]
