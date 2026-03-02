import argparse
import asyncio
import json

import questionary
import websockets
from websockets.exceptions import ConnectionClosed


async def start_publisher(host: str, port: int):
    uri = f"ws://{host}:{port}/ws-conn-pub"

    try:
        async with websockets.connect(uri) as websocket:
            print(f"Connected to Broker at {uri}")
            print("Type 'exit' or press Ctrl+C to quit.\n")

            while True:
                # Get topic
                topic = await questionary.text(
                    "Topic:",
                    instruction="Enter the channel name you want to publish to",
                ).ask_async()

                if topic is None or topic.lower() == "exit":
                    break

                if not topic.strip():
                    print("Topic cannot be empty.")
                    continue

                # Get message
                message = await questionary.text(
                    "Message:", instruction="Enter your message content"
                ).ask_async()

                if message is None or message.lower() == "exit":
                    break

                # Construct JSON payload & send
                payload = {"topic": topic.strip(), "text": message}
                try:
                    await websocket.send(json.dumps(payload))
                    print(f"Sent to [{topic.strip()}]")
                except ConnectionClosed:
                    print("\nConnection to Broker lost.")
                    break

    except (ConnectionRefusedError, OSError):
        print(f"Could not connect to Broker at {uri}.")
    except KeyboardInterrupt:
        print("\nPublisher shutting down.")
    finally:
        print("\nConnection closed. Goodbye!")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Distributed Pub/Sub Publisher")
    parser.add_argument("--host", type=str, default="localhost")
    parser.add_argument("--port", type=int, default=8242)

    args = parser.parse_args()

    try:
        asyncio.run(start_publisher(args.host, args.port))
    except KeyboardInterrupt:
        pass
