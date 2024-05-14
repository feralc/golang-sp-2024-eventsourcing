FROM golang:1.22.0 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM builder AS app_builder
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags '-extldflags "-static"' -o app
FROM alpine:3.15
ENV GOTRACEBACK=single
COPY --from=app_builder /app/app .
COPY --from=app_builder /app/.env .
RUN chmod a+x app
CMD ["./app", "$APP_COMMAND"]