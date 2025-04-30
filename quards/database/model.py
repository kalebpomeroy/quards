# db/models.py

import json
from contextlib import contextmanager
from datetime import datetime
from quards.database.db import get_connection


def insert_state(game_id, state_signature, state_obj):

    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO states (game_id, state_signature, state_json)
                VALUES (%s, %s, %s)
                ON CONFLICT DO NOTHING;
            """,
                (game_id, state_signature, json.dumps(state_obj)),
            )
            conn.commit()


def get_state(game_id, state_signature):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT game_id, game, state_json FROM states 
                WHERE state_signature = %s AND game_id = %s
                LIMIT 1;
            """,
                (state_signature, game_id),
            )
            row = cur.fetchone()

            if row:
                return {
                    "game_id": row[0],
                    "game": row[1],
                    "state": row[2],
                }
            return None


def insert_edge(game_id, parent_sig, name, params=None):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO edges (game_id, parent_signature, name, params, status)
                VALUES (%s, %s, %s, %s, 'OPEN')
                ON CONFLICT DO NOTHING;
            """,
                (game_id, parent_sig, name, params),
            )
            conn.commit()


def get_pending_edge():
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT parent_signature, name, params, id, game_id 
                FROM edges 
                WHERE status = 'OPEN' 
                LIMIT 1;
            """
            )
            row = cur.fetchone()

            if row:
                return {
                    "parent_signature": row[0],
                    "name": row[1],
                    "params": row[2],
                    "id": int(row[3]),
                    "game_id": row[4],
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
