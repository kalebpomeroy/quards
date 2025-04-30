from quards.evaluator.lorcana import game

LORCANA = "lorcana"


def execute(state, action, params):
    if action == "start":
        return game.start(state)


def get_empty_state():
    return {}
