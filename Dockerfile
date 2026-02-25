FROM golang:1.22-bookworm

RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    git \
    expect \
    && rm -rf /var/lib/apt/lists/*

RUN pip3 install --break-system-packages python-lsp-server

WORKDIR /app

COPY go.mod go.sum ./

ENV GOTOOLCHAIN=auto

RUN go mod download

COPY . .

RUN go build -o tron ./cmd/tron

ENV TERM=xterm-256color
ENV LINES=40
ENV COLUMNS=120

COPY test_tron.exp /app/

CMD ["expect", "-f", "test_tron.exp"]
