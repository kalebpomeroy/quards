from quards import turn_explorer
from quards.state_machine.database import model
from quards.state_machine.state import State
import sys

from quards.state_machine.database.db import run_sql

if __name__ == "__main__":

    # Ideally this in a unique way. This is the ID of the same and the seed
    seed = "my-new-game-uuid"

    state = State.from_id(seed, sys.argv[1])

    turn_explorer.explore(seed, state)

    next_turn = model.get_states_for_turn(seed, state.data["turn"] + 3)

    print(f"\tTotal ways to end the turn: {len(next_turn)}")

    print("Every universe has been explored")
