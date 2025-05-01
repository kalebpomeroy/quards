import json
import hashlib
from quards.database import model
from quards.evaluator import lorcana


class State:

    def __init__(self, game, seed, data):
        """
        Initialize a State object.

        :param game: what game/system is this should be using
        :param data: A dictionary representing the full state.
        """
        self.seed = seed
        self.data = data
        self.game = game

    def signature(self):
        """
        Creates a deterministic signature of the state for deduplication.

        Returns:
            A SHA-256 hex digest string.
        """
        serialized = json.dumps(self.data, sort_keys=True, separators=(",", ":"))
        # I'm not sure if we need the seed as part of the signature. This might be
        # overly protective, but is there a downside? More thinking needed.
        return "{}-{}".format(
            self.seed, hashlib.sha256(serialized.encode("utf-8")).hexdigest()
        )

    @classmethod
    def from_id(cls, seed, state_signature):
        """
        Loads a State object from a seed and signature

        :param seed: what seed state belongs to
        :param state_signature: The ID of the state
        :return: State instance
        """
        state_model = model.get_state(seed, state_signature)
        return State(state_model["game"], state_model["seed"], state_model["state"])

    @classmethod
    def new(cls, game, seed, data):
        """
        Create a State object in the database from a JSON-serializable dictionary.

        :param seed: The seed the state will belong to
        :param data: Dictionary representing the state.
        :return: State instance
        """
        state = State(game, seed, data)
        model.insert_state(state.seed, state.signature(), state.data)
        return state

    def get_actions(self):
        """
        For a state, get all of the actions that we could take from here,
        returns a list of tuples of action name plus a dict of params

        :return [ (name, params)... ]
        """
        if self.game == lorcana.LORCANA:
            return lorcana.get_actions(self.data)
