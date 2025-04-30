import random
from pathlib import Path
from quards.evaluator.lorcana import cards


class Deck:
    def __init__(self, deck_name):
        self.deck_name = deck_name
        self.deck = load_deck(deck_name)

    def shuffle(self, seed):
        rng = random.Random(seed)
        rng.shuffle(self.deck)


def load_deck(deck_name):

    deck = []
    path = Path(__file__).parent / f"data/decks/{deck_name}.dek"
    with open(path, "r") as f:
        for line in f:

            line = line.strip()
            if not line:
                continue
            count, text = line[0], line[2:]
            card = cards.CardIndex().get_by_title(text)
            deck.extend([card["Unique_ID"]] * int(count))

    return deck
