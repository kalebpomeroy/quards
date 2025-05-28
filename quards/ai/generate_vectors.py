from sentence_transformers import SentenceTransformer
from pprint import pprint
import json
import re

model = SentenceTransformer("nomic-ai/nomic-embed-text-v1", trust_remote_code=True)


def get_color_jargon(color):
    jargon = ""
    if "Amber" in color:
        jargon += "Amber (Yellow Y)"
    if "Amethyst" in color:
        jargon += "Amethyst (Purple P)"
    if "Ruby" in color:
        jargon += "Ruby (Red R)"
    if "Emerald" in color:
        jargon += "Emerald (Green G)"
    if "Sapphire" in color:
        jargon += "Sapphire (Blue U S)"
    if "Steel" in color:
        jargon += "Steel (Silver Grey S)"
    return jargon


def fix_text(text):
    text = text.replace("{w}", "willpower")
    text = text.replace("{s}", "strength")
    text = text.replace("{l}", "lore")
    text = text.replace("{i}", " ink")
    text = text.replace("{e}", "exert")
    text = text.replace("{n}", "inkable circles")

    if ":" not in text:
        return text

    return re.sub(r"^([^:\n]{1,50}):", "ability:", text, flags=re.MULTILINE)


def fix_name(name):
    if " - " not in name:
        return name
    name, version = name.split(" - ")
    return f"{name} (version: {version})"


def get_stats(card):
    if card["Type"] == "Location":
        return f"willpower (health, hp hit points, toughness): {card['Willpower']} Move Cost: {card['Move_Cost']} gains {card.get('Lore', 0)} lore a turn"

    if card["Type"] == "Character":
        return f"willpower (health, hp hit points, toughness): {card['Willpower']} Strength (attack, power, combat): {card['Strength']} quests for {card.get('Lore', 0)}"

    return ""


def format_card(card):

    return f" {fix_name(card["Name"])} {card['Cost']} Cost {get_color_jargon(card['Color'])} {card['Type']}.  lore. {get_stats(card)} Text: {fix_text(card.get('Body_Text', ""))}"


with open("data/cards.json") as f:
    cards = json.load(f)

texts = [format_card(card) for card in cards]
# [print(text, "\n") for text in texts]

vectors = model.encode(texts, batch_size=16, show_progress_bar=True)
embedded_cards = [
    {"name": cards[i]["Name"], "text": texts[i], "vector": vectors[i].tolist()}
    for i in range(len(cards))
]

with open("data/card_vectors.json", "w") as f:
    json.dump(embedded_cards, f)
