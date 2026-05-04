FROM golang:1.25.7 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/auth

FROM alpine:latest AS runtime
COPY --from=build /auth-service /auth-service
WORKDIR /app
EXPOSE 50051
CMD ["/auth-service"]
