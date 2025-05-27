from quards.state_machine import explorer
from quards.state_machine.state import State
from quards.state_machine.evaluator import lorcana

from quards.state_machine.database.db import run_sql_file, run_sql

if __name__ == "__main__":

    # For testing purposes, I'm just creating the database for scratch each time
    run_sql("DROP TABLE IF EXISTS states ;  DROP TABLE IF EXISTS edges;")
    run_sql_file("quards/state_machine/database/setup.sql")

    # Ideally this in a unique way. This is the ID of the same and the seed
    seed = "my-new-game-uuid"

    # Decks are stored on the filesystem in quards.evaluator.lorcana.data.decks
    # This method returns a dict of a state to start at
    state_data = lorcana.get_initial_state(seed, "yellow-test", "purple-test")
    state = State.new(lorcana.LORCANA, seed, state_data)

    # Eventually this would be daemonized to be infinite workers on all states
    # For testing, we're just exploring our current seed
    explorer.start_explore(seed, state)

    print("Every universe has been explored")
