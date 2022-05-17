FROM golang:1.18-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /app/k8s-simple-logs
EXPOSE 8080
ENTRYPOINT /app/k8s-simple-logs

