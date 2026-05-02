FROM golang:1.25.7 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /attachments-service ./cmd/attachments

FROM alpine:latest AS runtime
COPY --from=build /attachments-service /attachments-service
WORKDIR /app
EXPOSE 50054
CMD ["/attachments-service"]
