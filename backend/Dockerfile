# Build context is the repo root (see docker-compose*.yml and ci.yml): the
# builder stage copies backend/, the clipper stage copies tools/clip/.
FROM golang:alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hube ./cmd/api

# Bundle the clip extractor with the same npm ci + build.mjs esbuild
# invocation as `make build-clip`, producing /clip/dist/clip.cjs.
FROM node:22-alpine AS clipper
WORKDIR /clip
COPY tools/clip/package.json tools/clip/package-lock.json ./
RUN npm ci
COPY tools/clip/build.mjs tools/clip/clip.mjs ./
RUN node build.mjs

FROM alpine:3.21
RUN apk add --no-cache nodejs ca-certificates tzdata
WORKDIR /app
COPY --from=builder /hube .
COPY --from=clipper /clip/dist/clip.cjs /app/clip/clip.cjs
ENV CLIP_SCRIPT=/app/clip/clip.cjs
EXPOSE 8090
CMD ["./hube"]
