import requests


def generate_descriptions(game_state):
    payload = {
        "model": "mistral",
        "prompt": f"Describe the following game state five different ways:\n\n{game_state}",
    }
    response = requests.post("http://localhost:11434/api/generate", json=payload)
    return response.json()["response"]


import json
import random
import uuid

with open("cards.json") as f:
    cards = json.load(f)

card_ids = [card["Unique_ID"] for card in cards]


def random_card_ids(count):
    return random.sample(card_ids, min(count, len(card_ids)))


def make_battlefield_entry(card_id):
    return {"card": card_id, "dmg": 0, "exerted": False, "turn_effects": []}


def generate_state():
    state = {
        "zones": {
            "hands": {
                "player1": random_card_ids(random.randint(2, 7)),
                "player2": random_card_ids(random.randint(2, 7)),
            },
            "decks": {
                "player1": random_card_ids(random.randint(30, 60)),
                "player2": random_card_ids(random.randint(30, 60)),
            },
            "discard": {
                "player1": random_card_ids(random.randint(0, 3)),
                "player2": random_card_ids(random.randint(0, 3)),
            },
            "battlefield": {
                "player1": [
                    make_battlefield_entry(cid)
                    for cid in random_card_ids(random.randint(0, 3))
                ],
                "player2": [
                    make_battlefield_entry(cid)
                    for cid in random_card_ids(random.randint(0, 3))
                ],
            },
        },
        "lore": {"player1": random.randint(0, 20), "player2": random.randint(0, 20)},
        "inkwell": {"player1": random.randint(0, 10), "player2": random.randint(0, 10)},
        "ink_drops_available": random.randint(0, 2),
        "ink_available": random.randint(0, 10),
        "bag": [],
        "current_player": random.choice(["player1", "player2"]),
        "off_player": (
            "player2" if state.get("current_player") == "player1" else "player1"
        ),
        "turn": 0,
        "winner": "",
        "complete": False,
    }
    return state


def generate_n_states(n):
    return [generate_state() for _ in range(n)]


if __name__ == "__main__":
    n = 100  # change as needed
    states = generate_n_states(n)
    with open("generated_states.json", "w") as f:
        json.dump(states, f, indent=2)
