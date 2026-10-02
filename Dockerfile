FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/trade ./cmd/trade

FROM alpine:3.22

RUN adduser -D -u 10001 app
USER app

COPY --from=build /out/trade /usr/local/bin/trade

EXPOSE 8080

ENTRYPOINT ["trade"]
