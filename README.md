# Pubsub

Distributed publish-subscribe monorepo using websockets. The main server is
written in Go, but the Publisher and Subscriber are written in Python.

## Setup

```
just init
just run-tmux # alternatively;

just publisher
just subscriber
just broker
```
