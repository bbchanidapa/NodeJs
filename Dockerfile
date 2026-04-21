# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /server /server

USER nonroot:nonroot

EXPOSE 3000

ENTRYPOINT ["/server"]
