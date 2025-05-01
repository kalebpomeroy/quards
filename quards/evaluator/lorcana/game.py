def draw(state_data, player, count):
    drawn = state_data["zones"]["decks"][player][:count]
    state_data["zones"]["hands"][player].extend(drawn)
    del state_data["zones"]["decks"][player][:count]
    return state_data


def ink(state_data, player, card_id):
    state_data["ink_drops_available"] -= 1
    state_data["zones"]["ink"][player].append(card_id)
    state_data["zones"]["hands"][player].remove(card_id)

    return state_data


def pass_turn(state_data):
    state_data["turn"] += 1
    state_data["ink_drops_available"] = 1
    state_data["current_player"], state_data["off_player"] = (
        state_data["off_player"],
        state_data["current_player"],
    )

    # This is gonna get big. I think we need to set a flag
    # to say "turn change" so when diviner lists actions, we
    # create an actions if necessary during the end step and/or untap
    new_state = draw(state_data, state_data["current_player"], 1)
    return new_state


def did_i_win(state_data):
    # Set a fake victory condition:
    if len(state_data["zones"]["ink"][state_data["current_player"]]) > 1:
        state_data["complete"] = True
        state_data["winner"] = state_data["current_player"]

    return state_data
