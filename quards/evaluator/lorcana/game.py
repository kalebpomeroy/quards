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

    state_data["ink_drops_available"] = 1
    state_data["current_player"], state_data["off_player"] = (
        state_data["off_player"],
        state_data["current_player"],
    )
    new_state = draw(state_data, state_data["current_player"], 1)
    return new_state
