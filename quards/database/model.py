# db/models.py

import psycopg2
import json
from contextlib import contextmanager
from datetime import datetime


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


def insert_state(seed_id, state_signature, state_obj):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO states (seed_id, state_signature, state_json)
                VALUES (%s, %s, %s)
                ON CONFLICT DO NOTHING;
            """,
                (seed_id, state_signature, json.dumps(state_obj)),
            )
            conn.commit()


def insert_edge(parent_sig, name, params=None):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO edges (parent_signature, name, params, status)
                VALUES (%s, %s, %s, 'OPEN')
                ON CONFLICT DO NOTHING;
            """,
                (parent_sig, name, params),
            )
            conn.commit()


def get_pending_edge():
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT parent_signature, name, params, id FROM edges WHERE status = 'OPEN' LIMIT 1;
            """
            )
            row = cur.fetchone()

            if row:
                return {
                    "parent_signature": row[0],
                    "name": row[1],
                    "params": row[2],
                    "id": int(row[3]),
                }
            return None


def resolve_edge(action_id, child_sig):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE edges SET child_signature = %s, status = 'CLOSED' WHERE id = %s;
            """,
                (child_sig, action_id),
            )
            conn.commit()
