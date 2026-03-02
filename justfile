host := "localhost"
port := "8242"

default:
    @just --list

# --- Setup Tasks ---
[group('setup')]
init:
    uv sync
    cd broker && go mod tidy

# --- Execution Tasks ---
[group('execution')]
broker:
    cd broker && go run cmd/broker/main.go

[group('execution')]
publisher:
    uv run publisher/main.py --host {{host}} --port {{port}}

[group('execution')]
subscriber:
    uv run subscriber/main.py --host {{host}} --port {{port}}

[group('execution')]
run-tmux:
    @# Kill existing session if it exists + start anew
    tmux kill-session -t pubsub 2>/dev/null || true
    tmux new-session -d -s pubsub -n 'progs'

    @# Broker
    tmux send-keys -t pubsub:progs.0 'just broker' C-m

    @# Subscriber
    tmux split-window -h -t pubsub:progs
    tmux send-keys -t pubsub:progs.1 'just subscriber' C-m

    @# Publisher
    tmux split-window -v -t pubsub:progs.1
    tmux send-keys -t pubsub:progs.2 'just publisher' C-m

    @# Open nvim in second window
    tmux new-window -t pubsub:2 -n 'edit'
    tmux send-keys -t pubsub:edit 'nvim' C-m

    @# Open running progs
    tmux select-window -t pubsub:progs
    tmux attach-session -t pubsub

# --- Build & Maintenance ---
[group('build')]
build-broker:
    @mkdir -p bin
    cd broker && go build -o ../bin/broker cmd/broker/main.go

[group('build')]
clean:
    rm -rf broker/bin/
    rm -rf broker/info.log broker/error.log broker/broker.db

# --- Quality & Linting ---
[group('quality')]
format-py:
    uv run ruff format .

[group('quality')]
lint-py:
    uv run ruff check . --fix

[group('quality')]
check: format-py lint-py
