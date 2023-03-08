# # Build the wechatbot binary
FROM golang:1.20 as builder

WORKDIR /app

# Install upx for compress binary file
RUN apt update && apt install -y upx

# Copy the go source
COPY . .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build and compression
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o server main.go \
    && upx server

FROM frolvlad/alpine-glibc:alpine-3.17_glibc-2.34 as final

WORKDIR /app

# Install some software
RUN apk update && apk add --no-cache supervisor \
    && mkdir -p /app/config

# Copy the binary file from builder
COPY --from=builder /app/server .

ADD resources /app/resources
ADD static /app/static

ENTRYPOINT ["/app/server"]
CMD ["-config", "/app/config"]
