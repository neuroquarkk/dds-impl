FROM golang:1.26.2 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

ARG APP_NAME

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o service ./cmd/${APP_NAME}


FROM gcr.io/distroless/static-debian13 AS run
WORKDIR /app

COPY --from=builder /app/service ./service
USER nonroot:nonroot

ENTRYPOINT [ "./service" ]
