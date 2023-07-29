# This Dockerfile utilizes a multi-stage build
ARG ALPINE_VERSION=3.18

FROM golang:1.20-alpine$ALPINE_VERSION AS builder
ARG TARGETARCH
ARG TARGETOS

# Install necessary build tools
RUN apk --update add make bash curl git

# Switch workdir, otherwise we end up in /go (default)
WORKDIR /

# Copy everything into build container
COPY . .

# Build the application
RUN make build/$TARGETOS-$TARGETARCH

# Now in 2nd build stage
FROM library/alpine:$ALPINE_VERSION
ARG TARGETARCH
ARG TARGETOS

# SSL and quality-of-life tools
RUN apk --update add bash curl ca-certificates && update-ca-certificates

# Copy bin, tools, scripts, migrations
COPY --from=builder /build/snitch-server-$TARGETOS-$TARGETARCH /snitch-server

RUN chmod +x /snitch-server

EXPOSE 8080
EXPOSE 9090

CMD ["/snitch-server", "--debug"]