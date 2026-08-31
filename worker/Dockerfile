# The production worker lives in backend/cmd/worker.  Build it from the
# repository root so Render does not deploy the legacy worker stub.
FROM golang:1.21-alpine AS builder

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/partflow-worker ./cmd/worker

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /out/partflow-worker /app/partflow-worker
CMD ["/app/partflow-worker"]
