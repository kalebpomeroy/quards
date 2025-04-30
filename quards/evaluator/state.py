import json
import hashlib
from quards.database import model


class State:
    def __init__(self, seed_id, data):
        """
        Initialize a State object.

        :param data: A dictionary representing the full state.
        """
        self.seed_id = seed_id
        self.data = data

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
            self.seed_id, hashlib.sha256(serialized.encode("utf-8")).hexdigest()
        )

    @classmethod
    def from_json(cls, json_data):
        """
        Creates a State object from a JSON-serializable dictionary.

        :param json_data: Dictionary representing the state.
        :return: State instance
        """
        return cls(json_data)

    @classmethod
    def new(cls, seed_id, data):
        state = State(seed_id, data)
        model.insert_state(state.seed_id, state.signature(), state.data)
        return state
