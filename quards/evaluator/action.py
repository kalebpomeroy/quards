from quards.evaluator.state import State
from quards.database import model
from quards.evaluator import lorcana
import copy


# NOTE on terminology:
# Actions are simply edges that connect state. Most of the application we use
# the word action as it's more consistent with how a user things of the map.
# Within the context of the Action, we start being a little more explicit about
# edges and traversals using graph language
class Action:

    def __init__(self, id, seed, start_state_sig, name, params=None):
        """
        Represents an action that can transform a state.

        :params seed: The seed the action belongs to
        :params start_state_sig: The starting state signature
        :param name:  Name for this action, will be used in the evaluator
        :param params: A dict of options that will be used to perform the action
        :param id: the action if, if loading from the database
        """
        self.start_state_sig = start_state_sig
        self.name = name
        self.params = params
        self.id = id
        self.seed = seed
        self.state = None

    def resolve_edge(self, new_state):
        """
        Given the resulting state, mark an edge complete.

        params: State object to mark as complete
        """
        model.resolve_edge(self.id, new_state)

    @classmethod
    def new(cls, seed, start_state_sig, name, params, turn):
        """
        Create a new action

        :param seed The seed for the action
        :param start_state_sig The parent state signature
        :param name the Action name
        :param params a dict of options for the action
        :param turn that this will be executed
        """
        id = model.insert_edge(seed, start_state_sig, name, params, turn)

        action = Action(id, seed, start_state_sig, name, params)
        return action

    @classmethod
    def get_pending_edge(cls, seed, turn):
        """
        Get any one pending edge for our seed at the current turn

        :param seed the seed for the action
        :param turn int of the current depth
        """
        action_dict = model.get_pending_edge(seed, turn)
        if action_dict is None:
            return None

        return Action(
            action_dict["id"],
            action_dict["seed"],
            action_dict["parent_signature"],
            action_dict["name"],
            action_dict["params"],
        )

    def get_starting_state(self):
        """
        Lazy load the parent state for this action, caching it as necessary
        """
        if self.state:
            return self.state

        return State.from_id(self.seed, self.start_state_sig)

    def execute(self):
        """
        This executes the action based on the matching evaluator.
        Note: As a side effect, this creates a new signature db row

        returns new State object after the action has been applied
        """
        state = self.get_starting_state()

        if state.game == lorcana.LORCANA:
            new_state_data = lorcana.execute(
                copy.deepcopy(state.data), self.name, self.params
            )

        return State.new(state.game, state.seed, new_state_data)
