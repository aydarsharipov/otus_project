# Один Dockerfile для сервера и клиента. Что собирать, задаётся через APP (server или client).
FROM golang:1.24-alpine AS build

ARG APP=server
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/${APP}

FROM alpine:3.20

RUN adduser -D -u 10001 app
USER app

COPY --from=build /out/app /app

ENTRYPOINT ["/app"]
