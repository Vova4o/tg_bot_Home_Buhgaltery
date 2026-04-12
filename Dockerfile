FROM golang:1.22-bookworm AS builder

# Install tesseract dependencies
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    libtesseract-dev \
    tesseract-ocr-rus \
    tesseract-ocr-eng

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o expense-bot ./cmd/bot

FROM debian:bookworm-slim

# Install runtime dependencies for tesseract and CA certificates
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/expense-bot .

EXPOSE 8080
CMD ["./expense-bot"]
