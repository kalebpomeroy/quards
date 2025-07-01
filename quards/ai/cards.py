import json

cards = None
cards_by_name = None

with open("quards/ai/data/cards.json") as f:
    cards = json.load(f)
    cards_by_name = {card["Name"].lower(): card for card in cards}


def by_name(name):
    return cards_by_name.get(name.lower(), None)
