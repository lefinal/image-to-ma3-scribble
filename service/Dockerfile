FROM golang:1.24 as builder
WORKDIR /work

# Download dependencies.
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

# Perform the actual build.
COPY . .
RUN CGO_ENABLED=0 go build -a -o /app ./main.go

FROM debian

EXPOSE 8080
ENV POTRACE_FILENAME=/potrace/potrace
ENV HTTP_API_LISTEN_ADDR=:8080

WORKDIR /
COPY potrace-1.16.linux-x86_64 ./potrace
COPY --from=builder /app .

CMD ["/app"]
