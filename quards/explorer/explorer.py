from quards.evaluator.action import Action
from quards.evaluator.state import State


def take_action():

    action = Action.get_pending_edge()
    if action is None:
        return None

    print(f"Taking action: {action.name}")
    print("fake evaluator gave me a new state")
    print("creating a new state object")
    new_state = State.new("my-unique-seed", {"zones": [[], []]})

    print("created jobs based on the new state")

    print(f"resolving the previous job {action.id}")
    action.resolve_edge(new_state.signature())

    return action
