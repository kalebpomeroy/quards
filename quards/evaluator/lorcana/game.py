from quards.evaluator.lorcana import battlefield


def draw(state_data, player, count):
    drawn = state_data["zones"]["decks"][player][:count]
    state_data["zones"]["hands"][player].extend(drawn)
    del state_data["zones"]["decks"][player][:count]
    return state_data


def ink(state_data, player, card_id, tapped=False):
    state_data["ink_drops_available"] -= 1
    state_data["zones"]["ink"][player].append(card_id)
    state_data["zones"]["hands"][player].remove(card_id)
    if not tapped:
        state_data["ink_available"] += 1

    return state_data


def play(state_data, player, card_id, ink):
    state_data["zones"]["hands"][player].remove(card_id)

    state_data["zones"]["battlefield"][player].append(battlefield.add(card_id))

    state_data["ink_available"] -= ink

    return state_data


def pass_turn(state_data):
    if did_i_eat_the_cards(state_data):
        state_data["complete"] = True
        state_data["error"] = "pruned"

    state_data["turn"] += 1
    state_data["ink_drops_available"] = 1
    state_data["current_player"], state_data["off_player"] = (
        state_data["off_player"],
        state_data["current_player"],
    )

    battlefield.ready_for_turn(
        state_data["zones"]["battlefield"][state_data["current_player"]]
    )

    # This is gonna get big. I think we need to set a flag
    # to say "turn change" so when diviner lists actions, we
    # create an actions if necessary during the end step and/or untap
    return draw(state_data, state_data["current_player"], 1)


def did_i_win(state_data):
    # Set a fake victory condition:
    if len(state_data["zones"]["ink"][state_data["current_player"]]) == 5:
        state_data["complete"] = True
        state_data["winner"] = state_data["current_player"]

    return state_data


def did_i_eat_the_cards(state_data):
    """
    This function is called when a player is about to pass to see if there is
    an obvious "good move" that they should have made. This kills the whole
    branch, so only to be used with absolute certainty. Rules defined here become
    effectively loss conditions for the game

    - Did I not ink T1 or T2?
    """
    p1_ink = len(state_data["zones"]["ink"]["player1"])
    p2_ink = len(state_data["zones"]["ink"]["player2"])

    if state_data["turn"] == 1 and p1_ink == 0:
        print(f"Pruning this path. P1 passed T{state_data["turn"]} with {p1_ink}")
        return True
    if state_data["turn"] == 2 and p2_ink == 0:
        print(f"Pruning this path. P2 passed T{state_data["turn"]} with {p2_ink}")
        return True

    if state_data["turn"] == 3 and p1_ink == 1:
        print(f"Pruning this path. P1 passed T{state_data["turn"]} with {p1_ink}")
        return True
    if state_data["turn"] == 4 and p2_ink == 1:
        print(f"Pruning this path. P2 passed T{state_data["turn"]} with {p2_ink}")
        return True

    return False
