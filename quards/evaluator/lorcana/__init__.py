from quards.evaluator.lorcana import game
from quards.evaluator.lorcana.deck import Deck
from quards.evaluator.lorcana import game
from quards.evaluator.lorcana import diviner
from quards.database import model
import yaml
from pathlib import Path

LORCANA = "lorcana"


def get_actions(state_data):
    return diviner.get_actions(state_data)


def execute(state_data, action, params):

    after_action_state = _do_action(state_data, action, params)

    state_data = game.did_i_win(state_data)

    return after_action_state


def _do_action(state_data, action, params):
    if action == "start":
        return state_data

    if action == "ink":
        return game.ink(state_data, **params)

    if action == "pass":
        return game.pass_turn(state_data)


def get_initial_state(seed, deck1, deck2):
    path = Path(__file__).parent / "data/initial_state.yaml"
    with open(path, "r") as f:
        state_data = yaml.safe_load(f)

    player1_deck = Deck(deck1)
    player2_deck = Deck(deck2)

    player1_deck.shuffle(seed)
    player2_deck.shuffle(seed)

    state_data["zones"]["decks"][state_data["current_player"]] = player1_deck.deck
    state_data["zones"]["decks"][state_data["off_player"]] = player2_deck.deck

    state_data = game.draw(state_data, state_data["current_player"], 7)
    state_data = game.draw(state_data, state_data["off_player"], 7)

    return state_data


def get_turn_summary(seed, turn):

    all_states = model.get_states_for_turn(seed, turn)

    print(f"\tTotal possibilities: {len(all_states)}")
    print(f"\tTotal paths to victory: {len([s for s in all_states if s['complete']])}")

    return True
