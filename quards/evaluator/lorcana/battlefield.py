def add(card_id):

    # TODO: Load the card and add logic like if card[Abilities]
    # includes rush, can challenge this turn is True, etc

    return {
        "id": card_id,
        "tapped": False,
        "can_quest": False,
        "can_challenge": False,
        "damage": 0,
        "location": None,
    }


def ready_for_turn(players_battlefield):
    for obj in players_battlefield:
        obj["tapped"] = False
        obj["can_quest"] = True
        obj["can_challenge"] = True

        # Do stuff for card type, text checks, etc
