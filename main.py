from quards.explorer import explorer
from quards.evaluator.state import State
from quards.evaluator.action import Action
from quards.evaluator import lorcana

from quards.database.db import run_sql_file, run_sql


if __name__ == "__main__":

    run_sql("DROP TABLE IF EXISTS states ;  DROP TABLE IF EXISTS edges;")
    run_sql_file("quards/database/setup.sql")

    # This is a good point of abstraction. This is where the "game" starts. Right
    # now it's just configured to have a single game

    game_id = "my-new-game-uuid"
    game = "lorcana"

    state_json = lorcana.get_initial_state(game_id, "yellow-test", "purple-test")
    state = State.new(game, game_id, state_json)

    Action.new(game_id, state.signature(), "start")

    while True:
        if explorer.take_action() is None:
            break

    print("nothing was left so we're done.")
