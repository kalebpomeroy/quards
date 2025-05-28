from sentence_transformers import SentenceTransformer
from pprint import pprint
import json
import numpy as np
from numpy.linalg import norm
from util import timed


@timed("Loading card database")
def load_cards():
    with open("data/card_vectors.json") as f:
        return json.load(f)


@timed("Loading embedding model")
def load_model():
    return SentenceTransformer("nomic-ai/nomic-embed-text-v1", trust_remote_code=True)


@timed("Encoding query")
def encode_query(model, text):
    return model.encode(text)


@timed("Running search")
def search(query_vector, card_db, top_k=5):
    scored = [
        (cosine_similarity(query_vector, card["vector"]), card) for card in card_db
    ]
    return sorted(scored, reverse=True)[:top_k]


def cosine_similarity(a, b):
    return np.dot(a, b) / (norm(a) * norm(b))


card_db = load_cards()
model = load_model()
query = encode_query(
    model,
    """Return cards that meet the following criteria:

- Color: Ruby (Red)
- High Strength (6 or greater)
- Preferably Characters (not Actions or Songs)
- Focus on raw attack power (not buffs, debuffs, or conditional effects)

Ignore cards that:
- Only grant temporary Strength
- Only reduce opponent Strength
- Are not Characters

Return full card text with Strength and Color clearly shown.""",
)
results = search(query, card_db)

for v, card in results:
    print(f"{card['name']} (score: {v:.3f})")
    print(card["text"])
    print()
