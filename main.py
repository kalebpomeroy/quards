from quards.explorer import explorer
from quards.evaluator.state import State
from quards.evaluator.action import Action

from quards.database.model import get_connection


def run_sql_file(path):
    with open(path, "r") as f:
        sql = f.read()
    with get_connection() as conn:
        with conn.cursor() as cur:
            # TESTING ONLY
            cur.execute("DROP TABLE IF EXISTS states ;  DROP TABLE IF EXISTS edges;")
            cur.execute(sql)
        conn.commit()


if __name__ == "__main__":

    run_sql_file("quards/database/setup.sql")

    state = State.new("my-unique-seed", {"zones": []})

    Action.new(state.signature(), "start")

    while True:
        if explorer.take_action() is None:
            break

    print("nothing was left so we're done.")
