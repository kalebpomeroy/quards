# db/models.py

import json
from contextlib import contextmanager
from datetime import datetime
from quards.database.db import get_connection


def insert_state(seed, state_signature, state_obj):

    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO states (seed, state_signature, state_json)
                VALUES (%s, %s, %s)
                ON CONFLICT DO NOTHING;
            """,
                (seed, state_signature, json.dumps(state_obj)),
            )
            conn.commit()


def get_state(seed, state_signature):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT seed, game, state_json FROM states 
                WHERE state_signature = %s AND seed = %s
                LIMIT 1;
            """,
                (state_signature, seed),
            )
            row = cur.fetchone()

            if row:
                return {
                    "seed": row[0],
                    "game": row[1],
                    "state": row[2],
                }
            return None


def insert_edge(seed, parent_sig, name, params, turn):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO edges (seed, parent_signature, name, params, turn, status)
                VALUES (%s, %s, %s, %s, %s, 'OPEN')
                ON CONFLICT DO NOTHING
                RETURNING id;
            """,
                (seed, parent_sig, name, json.dumps(params), turn),
            )
            result = cur.fetchone()
            conn.commit()
            return result[0] if result else None


def get_pending_edge(seed, turn):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT parent_signature, name, params, id, seed 
                FROM edges 
                WHERE status = 'OPEN' AND turn = %s AND seed = %s
                LIMIT 1;
            """,
                (turn, seed),
            )
            row = cur.fetchone()

            if row:
                return {
                    "parent_signature": row[0],
                    "name": row[1],
                    "params": row[2],
                    "id": int(row[3]),
                    "seed": row[4],
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
