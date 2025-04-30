# evaluator/action.py

from quards.database import model


class Action:
    def __init__(self, start_state_sig, name, params=None, id=None):
        """
        Represents an action that can transform a state.

        :param action_id: Unique identifier for this action (string or UUID)
        :param name: Human-readable name for this action
        :param apply_fn: A function taking (state) -> new state
        """
        self.start_state_sig = start_state_sig
        self.name = name
        self.params = params
        self.id = id

    def apply(self, state):
        """
        Apply this action to a given State object and return a new State.
        """
        return self.apply_fn(state)

    def __repr__(self):
        return f"Action({self.name})"

    def resolve_edge(self, new_state):
        model.resolve_edge(self.id, new_state)

    @classmethod
    def new(cls, start_state_sig, name, params=None):
        action = Action(start_state_sig, name, params)

        # TODO: This doesn't set the ID. It doesn't matter yet, but I might
        #       need this later if it's important.
        model.insert_edge(action.start_state_sig, action.name, action.params)
        return action

    @classmethod
    def get_pending_edge(cls):
        action_dict = model.get_pending_edge()
        if action_dict is None:
            return None

        return Action(
            action_dict["parent_signature"],
            action_dict["name"],
            action_dict["params"],
            action_dict["id"],
        )
