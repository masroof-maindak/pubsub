import argparse
import asyncio
import json

import questionary
import websockets
from websockets.exceptions import ConnectionClosed


async def message_listener(websocket):
    """Background task to continuously listen for messages from the broker."""
    try:
        async for message in websocket:
            print(f"\n[Incoming Message]: {message}")
            print("> ", end="", flush=True)
    except ConnectionClosed:
        print("\nConnection to Broker closed by server.")


async def start_subscriber(host: str, port: int):
    uri = f"ws://{host}:{port}/ws-conn-sub"

    try:
        async with websockets.connect(uri) as websocket:
            print(f"Connected to Broker as Subscriber at {uri}")

            # Start background listener for incoming messages
            listener_task = asyncio.create_task(message_listener(websocket))

            while True:
                # Ask action
                action = await questionary.select(
                    "What would you like to do?",
                    choices=["SUBSCRIBE", "UNSUBSCRIBE", "Exit"],
                ).ask_async()

                if action == "Exit" or action is None:
                    break

                # Ask topic
                topic = await questionary.text(
                    f"Enter topic to {action.lower()}:"
                ).ask_async()

                if not topic or not topic.strip():
                    print("Topic cannot be empty.")
                    continue

                # Construct JSON payload & send
                payload = {"action": action, "topic": topic.strip()}
                try:
                    await websocket.send(json.dumps(payload))
                    print(f"Command sent: {action} {topic.strip()}")
                except ConnectionClosed:
                    break

            # Cleanup
            listener_task.cancel()
            try:
                await listener_task
            except asyncio.CancelledError:
                pass

    except (ConnectionRefusedError, OSError):
        print(f"Could not connect to Broker at {uri}.")
    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        print("\nSubscriber shutting down.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Distributed Pub/Sub Subscriber")
    parser.add_argument("--host", type=str, default="localhost")
    parser.add_argument("--port", type=int, default=8242)

    args = parser.parse_args()

    try:
        asyncio.run(start_subscriber(args.host, args.port))
    except KeyboardInterrupt:
        pass
