from quards.evaluator.action import Action
from quards.evaluator.state import State


def take_action():

    action = Action.get_pending_edge()
    if action is None:
        return None

    print(f"\tTaking action: {action.name}...")
    new_state, actions = action.execute()

    print(f"\t\tTotal of {len(actions)} to be added...")
    for a in actions:
        Action.new(action.game_id, new_state.signature(), a["name"], a["params"])

    print(f"\t\t\tResolving the action")
    action.resolve_edge(new_state.signature())

    return action
