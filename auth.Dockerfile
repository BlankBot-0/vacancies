FROM golang:1.22.2-alpine3.18 as builder
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go clean --modcache
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o cmd/auth_fetcher/bin/main ./cmd/auth_fetcher/

FROM chromedp/headless-shell
WORKDIR /vacancies
COPY /.env .
COPY /auth_fetcher/config.yaml ./config/
COPY --from=builder /app/cmd/auth_fetcher/bin/main .
EXPOSE 8080
CMD ["./main"]