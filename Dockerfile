FROM golang:latest as build

WORKDIR /app

COPY go.mod .
# COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /chatapp ./cmd/chatapp/main.go
 
FROM alpine:latest as run

WORKDIR /app

COPY --from=build /chatapp /chatapp

EXPOSE 8080
CMD ["/app/chatapp"]