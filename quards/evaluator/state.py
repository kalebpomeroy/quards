import json
import hashlib
from quards.database import model


class State:
    def __init__(self, game, game_id, data):
        """
        Initialize a State object.

        :param data: A dictionary representing the full state.
        """
        self.game_id = game_id
        self.data = data
        self.game = game

    def to_json(self):
        """
        Returns the state as a JSON-serializable dictionary.
        """
        return self.data

    def signature(self):
        """
        Creates a deterministic signature of the state for deduplication.

        Returns:
            A SHA-256 hex digest string.
        """
        serialized = json.dumps(self.data, sort_keys=True, separators=(",", ":"))
        return "{}-{}".format(
            self.game_id, hashlib.sha256(serialized.encode("utf-8")).hexdigest()
        )

    @classmethod
    def from_id(cls, game_id, state_signature):
        """
        Loads a State object from a Signature

        :param game_id: what game ID the state belongs to
        :param state_signature: The ID of the state
        :return: State instance
        """
        state_model = model.get_state(game_id, state_signature)
        return State(state_model["game"], state_model["game_id"], state_model["state"])

    @classmethod
    def new(cls, game, game_id, data):
        """
        Create a State object in the database from a JSON-serializable dictionary.

        :param game_id: The game ID the state will belong to
        :param data: Dictionary representing the state.
        :return: State instance
        """
        state = State(game, game_id, data)
        model.insert_state(state.game_id, state.signature(), state.data)
        return state
