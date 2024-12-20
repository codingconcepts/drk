# Compile
FROM golang:1.23.4-alpine AS build
WORKDIR /src
COPY . /src

# Arguments
ARG version

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=$version" -o /app drk.go

# Build
FROM alpine
COPY --from=build app .
COPY AmazonRootCA1.pem .
ENTRYPOINT [ "./app" ]
