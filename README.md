# Pubsub

Distributed publish-subscribe monorepo using websockets. The main server is
written in Go, but the Publisher and Subscriber are written in Python.

## Setup

```
just init
just run-tmux # alternatively;

just broker
just publisher
just subscriber
```

## Features

- Trivial thread-safety via mutex-controlled map of topics (string) to a slice
  of websocket connection object pointers
- Robust error handling, logging and idiomatic channel usage
- Unsubscribe functionality added
- SQLite for topic persistence and painless backups

## Future Scalability

- 'Topic sharding' to use different mutexes for 'ranges' of topics
- Can further distribute these across different broker server instances
- Easy to detect downages via ping-ponging fails; and a separate 'monitoring'
  server can 'take over' as the lead broker should the broker
  abruptly/unexpectedly fail
- If IO starts bottlenecking server performance, we can trivially move database
  operations to a separate device
- SQLite -> Postgres migration is easy and cloud-hostable
