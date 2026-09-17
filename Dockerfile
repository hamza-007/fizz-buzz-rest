FROM golang:1.22-alpine AS build

RUN apk add --update --no-cache ca-certificates

WORKDIR /src

COPY . .

RUN go build -ldflags="-s -w" -trimpath -o fizzbuzz

FROM alpine

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

COPY --from=build /src/fizzbuzz /app/fizzbuzz

# An exploit in the server must not land as root in the container.
RUN adduser -D -u 65532 app
USER 65532:65532

ENV ENVIRONMENT=production
ENV APP_PORT=8080

EXPOSE 8080

CMD ["/app/fizzbuzz","serve"]
