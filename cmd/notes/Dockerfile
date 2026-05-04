FROM golang:1.25.7 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /notes-service ./cmd/notes

FROM alpine:latest AS runtime
COPY --from=build /notes-service /notes-service
WORKDIR /app
EXPOSE 50052
CMD ["/notes-service"]
