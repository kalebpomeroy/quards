from quards.evaluator.lorcana import cards


def get_actions(state_data):

    if state_data["complete"]:
        return []

    # If there's anything in the bag, all we can do is take stuff out of the bad
    if len(state_data["bag"]) > 0:
        return state_data["bag"]

    player = state_data["current_player"]

    # Assuming the bag is empty, we can pass, or take more action
    actions = [("pass", None)]

    for card_id in state_data["zones"]["hands"][player]:
        card = cards.CardIndex().get_by_id(card_id)

        if state_data["ink_drops_available"] > 0:
            if card["Inkable"]:
                actions.append(("ink", {"card_id": card_id, "player": player}))

        if card["Cost"] <= state_data["ink_available"]:
            actions.append(
                ("play", {"card_id": card_id, "player": player, "ink": card["Cost"]})
            )

    return actions
