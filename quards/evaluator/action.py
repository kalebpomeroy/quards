from quards.evaluator.state import State
from quards.database import model
from quards.evaluator import lorcana
import copy


class Action:

    def __init__(self, game_id, start_state_sig, name, params=None, id=None):
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
        self.game_id = game_id
        self.state = None

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
    def new(cls, game_id, start_state_sig, name, params=None):
        action = Action(game_id, start_state_sig, name, params)

        # TODO: This doesn't set the ID. It doesn't matter yet, but I might
        #       need this later if it's important.
        model.insert_edge(game_id, action.start_state_sig, action.name, action.params)
        return action

    @classmethod
    def get_pending_edge(cls):
        action_dict = model.get_pending_edge()
        if action_dict is None:
            return None

        return Action(
            action_dict["game_id"],
            action_dict["parent_signature"],
            action_dict["name"],
            action_dict["params"],
            action_dict["id"],
        )

    def get_state(self):
        if self.state:
            return self.state

        return State.from_id(self.game_id, self.start_state_sig)

    def execute(self):
        state = self.get_state()

        if state.game == lorcana.LORCANA:
            new_state_data, actions = lorcana.execute(
                copy.deepcopy(state.data), self.name, self.params
            )

        return State.new(state.game, state.game_id, new_state_data), actions
