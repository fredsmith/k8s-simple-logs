FROM golang:1.24-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY VERSION ./

ARG VERSION=dev
RUN go build -ldflags "-X main.version=${VERSION}" -o /app/k8s-simple-logs
EXPOSE 8080
ENTRYPOINT /app/k8s-simple-logs

