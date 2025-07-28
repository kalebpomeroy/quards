import psycopg2
from contextlib import contextmanager


@contextmanager
def get_connection():
    conn = psycopg2.connect(
        dbname="explorer_db",
        user="explorer",
        password="explorer",
        host="localhost",
        port=5432,
    )
    try:
        yield conn
    finally:
        conn.close()


def run_sql_file(path):
    with open(path, "r") as f:
        sql = f.read()
    run_sql(sql)


def run_sql(sql):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(sql)
        conn.commit()
