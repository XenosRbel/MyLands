# Multi-stage build: web client → Go server → minimal runtime image.
# NOTE: written for Docker; not executed in the authoring environment (no Docker daemon there).
FROM node:24-alpine AS client
WORKDIR /src/client
COPY client/package.json client/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY client/ ./
RUN npm run build

FROM golang:1.27-alpine AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/packtool ./cmd/packtool

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=server /out/server /out/packtool /app/
COPY packs /app/packs
COPY --from=client /src/client/dist /app/client
ENV RF_ENV=production RF_HTTP_ADDR=:8080 RF_PACKS_DIR=/app/packs RF_CLIENT_DIR=/app/client
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/app/server"]
