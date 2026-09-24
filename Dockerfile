# syntax=docker/dockerfile:1

# ============================================================
# 1. React frontend
# ============================================================

FROM node:24.15.0 AS frontend-builder

WORKDIR /app/web

# Install dependencies separately so Docker can cache this layer
COPY web/package*.json ./

RUN npm ci

COPY web/ .

RUN npm run build


# ============================================================
# 2. Go base
# ============================================================

FROM golang:1.26.3 AS go-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


# ============================================================
# 3. Go development builder
#
# Used locally.
# BuildKit cache keeps downloaded modules and compiled packages.
# ============================================================

FROM go-base AS builder-local

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-s -w" \
    -o /api \
    ./cmd/api


# ============================================================
# 4. Go production builder
#
# Used by Google Cloud / CI.
# No cache mounts.
# ============================================================

FROM go-base AS builder

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-s -w" \
    -o /api \
    ./cmd/api




# ============================================================
# 5. Local runtime
#
# Same final environment, but takes Go binary from
# the cached local builder.
# ============================================================

FROM nginx:alpine AS development

RUN apk add --no-cache gettext ca-certificates

WORKDIR /app

# Go API
COPY --from=builder-local /api .

# React
COPY --from=frontend-builder \
    /app/web/dist \
    /usr/share/nginx/html

# Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/configfile.template

# Startup script
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENV PORT=8080

EXPOSE 8080

CMD ["/entrypoint.sh"]

# ============================================================
# 6. Production runtime
#
# Uses the production Go builder.
# ============================================================

FROM nginx:alpine AS production

RUN apk add --no-cache gettext ca-certificates

WORKDIR /app

# Go API
COPY --from=builder /api .

# React
COPY --from=frontend-builder \
    /app/web/dist \
    /usr/share/nginx/html

# Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/configfile.template

# Startup script
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENV PORT=8080

EXPOSE 8080

CMD ["/entrypoint.sh"]
