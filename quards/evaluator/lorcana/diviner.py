from quards.evaluator.lorcana import cards


def get_actions(state_data):

    if state_data["complete"]:
        return []

    player = state_data["current_player"]

    actions = [("pass", None)]
    if state_data["ink_drops_available"] > 0:
        for card_id in state_data["zones"]["hands"][player]:
            card = cards.CardIndex().get_by_id(card_id)
            if card["Inkable"]:
                actions.append(("ink", {"card_id": card_id, "player": player}))

    return actions
