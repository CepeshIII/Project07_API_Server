# --- Stage 1: Build React Frontend ---
FROM node:24.15.0 AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ .

RUN npm run build



# --- Stage 2: Build Go Backend ---
FROM golang:1.26.3 AS builder
WORKDIR /app

# Copy dependency files first to cache module downloads
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the rest of the source code
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api


# --- Stage 3: Production Runtime ---
FROM nginx:alpine

# Install gettext so envsubst can inject the PORT variable
RUN apk add --no-cache gettext ca-certificates

WORKDIR /app

# Copy built Go binary from backend builder
COPY --from=builder /app/api .

# Copy built React files from frontend builder to Nginx public directory
COPY --from=frontend-builder /app/web/dist /usr/share/nginx/html

# Copy Nginx config template and entrypoint script from project root
COPY nginx.conf /etc/nginx/conf.d/configfile.template
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Cloud Run dynamic port mapping
ENV PORT=8080
EXPOSE 8080

CMD ["/entrypoint.sh"]
