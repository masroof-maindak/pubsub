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
    tmux new-session -d -s pubsub -n 'programs'

    @# Broker
    tmux send-keys -t pubsub:programs.0 'just broker' C-m

    @# Subscriber
    tmux split-window -h -t pubsub:programs
    tmux send-keys -t pubsub:programs.1 'sleep 1 && just subscriber' C-m

    @# Publisher
    tmux split-window -v -t pubsub:programs.1
    tmux send-keys -t pubsub:programs.2 'sleep 1 && just publisher' C-m

    @# Open nvim in second window
    tmux new-window -t pubsub:2 -n 'nvim'
    tmux send-keys -t pubsub:editor 'nvim .' C-m

    @# Open running programs
    tmux select-window -t pubsub:programs
    tmux attach-session -t pubsub

# --- Build & Maintenance ---
[group('build')]
build-broker:
    @mkdir -p bin
    cd broker && go build -o ../bin/broker cmd/broker/main.go

[group('build')]
clean:
    rm -rf bin/
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
