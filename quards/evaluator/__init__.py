from quards.evaluator import lorcana

evaluators = [lorcana]


def get_evaluator(game):
    if game == lorcana.LORCANA:
        return lorcana

    # Other systems can be created here, so long as they honor to the contract

    # def execute(state.data, action, params)
    # def get_actions(state.data)
