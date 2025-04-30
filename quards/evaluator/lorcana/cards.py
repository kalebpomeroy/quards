import json
from pathlib import Path


class CardIndex:

    _instance = None

    def __new__(cls, path="data/cards.json"):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialize(path)
        return cls._instance

    def _initialize(self, path):
        with open(Path(__file__).parent / path, "r") as f:
            self.cards = json.load(f)

        self.by_title = {card["Name"].lower(): card for card in self.cards}
        self.by_id = {card["Unique_ID"].lower(): card for card in self.cards}

    def get_by_title(self, name):
        return self.by_title.get(name.lower())

    def get_by_id(self, card_id):
        return self.by_id.get(card_id.lower())
