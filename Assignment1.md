# **Programming Assignment 1: A Distribute Publish-Subscribe System in C++**

February 23, 2026

## 1. Objective

The goal of this assignment is to gain practical experience with distributed
system architectures, low-level socket programming, and concurrency in C++. You
will implement a simplified version of a Publish-Subscribe (Pub/Sub) messaging
system, a common architectural style as discussed in Chapter 2.1.3 of the course
text.

Your system will consist of three separate C++ applications: a central
**Broker** (the server), **Publishers** (clients that send messages), and
**Subscribers** (clients that receive messages). This project will help you
understand:

- The client-server and publish-subscribe architectural patterns.
- Network communication using TCP sockets (POSIX sockets or Winsock).
- Server-side concurrency using C++ standard libraries (`<thread>`, `<mutex>`).
- The importance of defining and implementing a clear communication protocol.
- Practical challenges related to distributed system design goals like
  dependability and scalability.

## 2. System Components & Requirements

You will create three separate executables: `broker`, `publisher`, and
`subscriber` .

### 2.1 The Broker (Server)

The broker is the central message router. It must be a long-running process that
manages the entire system.

1. **Listen for Connections:** The broker must start up, create a TCP socket,
   bind it to a specific port provided as a command-line argument, and listen
   for incoming connections.
2. **Handle Multiple Clients Concurrently:** For every client that connects (via
   `accept()`), the broker must spawn a new **worker thread** (`std::thread`) to
   handle all communication with that specific client. This allows the main
   thread to immediately go back to listening for new clients.
3. **Manage Topics & Subscriptions:** The broker must maintain a shared data
   structure that maps topic strings to a list of subscribers. A good choice is
   `std::map<std::string, std::vector<int>>`, where the `int` is the client’s
   socket descriptor.
4. **Thread Safety:** Because multiple worker threads will need to access the
   subscription map (e.g., one thread adds a subscriber while another reads the
   list to publish a message), this shared resource **must be protected** from
   race conditions using a `std::mutex` . Use `std::lock_guard` or
   `std::unique_lock` for safe, exception-proof locking.
5. **Process Client Commands:** Each worker thread will read data from its
   client socket and parse commands according to the protocol defined in
   Section 3.
   - When a `SUBSCRIBE` command is received, add the client’s socket descriptor
     to the list for the given topic.
   - When a `PUBLISH` command is received, iterate through the list of
     subscribers for that topic and send the message to each one.
6. **Handle Disconnections:** If a client disconnects (e.g., `recv()` returns 0
   or -1), the corresponding worker thread must clean up. This includes closing
   the client’s socket and, critically, removing that client’s socket descriptor
   from all topic lists it was subscribed to. This cleanup must also be
   protected by the mutex.

### 2.2 The Publisher (Client)

The publisher is a command-line application that sends messages.

1. **Connect to the Broker:** Take the broker’s IP address and port as
   command-line arguments and establish a TCP connection.
2. **Send Messages:** In a loop, prompt the user for a topic and a message.
   Format this information according to the protocol and send it to the broker.

### 2.3 The Subscriber (Client)

The subscriber is a command-line application that subscribes to topics and
receives messages.

1. **Connect to the Broker:** Take the broker’s IP address and port as
   command-line arguments.
2. **Subscribe to Topics:** Take a list of topics to subscribe to from the
   command-line arguments. After connecting, send a `SUBSCRIBE` command for each
   topic.
3. **Receive Messages:** Enter an infinite loop to listen for incoming messages
   from the broker. When a message is received, parse it and print it to the
   console in a user-friendly format (e.g., `[Topic:`
   `weather] It is now raining.` ).

## 3. Communication Protocol

To ensure interoperability, all components must adhere to the following simple,
text-based protocol. Every message sent over a socket must be a single line of
text terminated by a newline character ( `’` \_\_ `n’` ).

**Subscriber to Broker (Subscription Request):** `SUBSCRIBE <topic_name>`
Example: `SUBSCRIBE sports`

**Publisher to Broker (Publish Message):**
`PUBLISH <topic_name> <message content>` Example:
`PUBLISH sports Final score: Team A 10, Team B 5`

**Broker to Subscriber (Forwarded Message):**
`MESSAGE <topic_name> <message content>` Example:
`MESSAGE sports Final score: Team A 10, Team B 5`

## 4. What to Submit

Submit a single `.zip` file containing:

1.  All source code files ( `.cpp` and `.h` ).
2.  Your build file ( `Makefile` or `CMakeLists.txt` ).
3.  A `README.md` file that includes:
    - Your name and student ID.
    - Clear, step-by-step instructions on how to build and run your three
      programs on your chosen platform (e.g., Linux or Windows).
    - A brief description of your design choices, especially how you implemented
      thread safety for the subscription map.
    - A short discussion (2-3 paragraphs) on how your implementation relates to
      two of the **design goals** from Chapter 1.2 (e.g., Scalability,
      Dependability, Openness). For example, what are the scalability
      bottlenecks in your centralized broker design? How could you improve its
      dependability if the broker crashes?

## 5. Bonus Features (Up to 10 extra points)

Implement one or more of the following advanced features for bonus points.
Clearly document which features you implemented in your README.

- [x] **Unsubscribe:** Implement an `UNSUBSCRIBE <topic_name>` command.
- [x] **Message History:** Implement a feature where a new subscriber can
      request the last known message for a topic upon subscribing. The broker
      would need to store the last message for each topic.
- [ ] **Wildcard Subscriptions:** Allow a subscriber to subscribe to topics
      using a wildcard, e.g., `SUBSCRIBE sports.*`, which would match topics
      like `sports.football` and `sports.basketball` .
