from quards.action import Action

from quards.evaluator import get_evaluator

MAX_DEPTH = 5


def start_turn_explore(seed, state):

    # Given a state create all actions from this state on.
    this_turn = state.data["turn"]
    for name, params in state.get_actions():
        Action.new(seed, state.signature(), name, params, this_turn)

    while True:
        # Get any action at our current turn for this game. We do not
        # grab actions for future turns, ensuring we process every possibility
        # for a given depth before moving deeper
        action = Action.get_pending_edge(seed, this_turn)

        # If there aren't any actions left at this level, return False so we
        # know we did nothing and can increase our depth and process the next turn
        if action is None:
            break

        # Create a new state representing the world after that action
        state = action.execute()

        # The current action is be resolved to point at the newly generated state
        action.resolve_edge(state)

        # If the state caused a turn change (usually a pass action) we don't want to go deeper
        if this_turn == state.data["turn"]:

            # Given a state, what is every possible action we could take and add it as
            # an action at the appropriate depth.
            for name, params in state.get_actions():
                Action.new(seed, state.signature(), name, params, state.data["turn"])
