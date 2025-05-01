from quards.evaluator.action import Action


MAX_DEPTH = 10


def start_explore(seed, state):

    # Given a seed and state create an action to kick of the game. This is
    # mostly a passthrough that creates the next turn of actions
    Action.new(seed, state.signature(), "start", None, 0)

    # By default we'll process actions at 0 (see above)
    depth = 0

    while True:
        # Take a random action for our game at the current working depth
        action_taken = take_action(seed, depth)

        # If we didn't do anything, we should increment the depth
        if not action_taken:
            depth += 1
            print(f"We delve deeper: {depth}")

        # Players can pass up to (60-7)*2 times. Those branches and many others
        # are quite uninteresting, so we're only going to go so far
        if depth > MAX_DEPTH:
            return None


def take_action(seed, depth):

    # Get any action at our current depth for this game. We do not
    # grab actions for future turns, ensuring we process every possibility
    # for a given depth before moving deeper
    action = Action.get_pending_edge(seed, depth)

    # If there aren't any actions left at this level, return False so we
    # know we did nothing and can increase our depth and process the next turn
    if action is None:
        return False

    # Create a new state representing the world after that action
    state = action.execute()

    # The current action is be resolved to point at the newly generated state
    action.resolve_edge(state.signature())

    # The next set of actions will be placed at this depth. This is the same
    # as the current_depth, with the exceptions of the pass and start actions
    # which increments turn, pushing all new actions into the future.
    next_depth = state.data["turn"]

    # Given a state, what is every possible action we could take and add it as
    # an action at the appropriate depth. Note, we can add actions past the
    # maximum depth. Not going deeper is a concern of the explorer func above
    for name, params in state.get_actions():
        Action.new(seed, state.signature(), name, params, next_depth)

    return True
