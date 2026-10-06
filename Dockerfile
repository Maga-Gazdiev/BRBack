FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app ./cmd/app
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /app /usr/local/bin/m96
COPY --chown=app:app data/content.json /app/data/content.json
RUN mkdir /app/data/uploads && chown -R app:app /app/data
USER app
ENV HTTP_ADDR=:8080 DATA_FILE=/app/data/content.json UPLOAD_DIR=/app/data/uploads
EXPOSE 8080
CMD ["m96"]
