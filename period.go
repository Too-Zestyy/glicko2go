package glicko2go

func PeriodCalculatorWithSettings(settings Glicko2AlgorithmSettings) func(
	players map[int]Glicko2Player,
	matches []Glicko2MatchByID) (map[int]Glicko2Player, error) {

	newPlayerUpdater := NewPlayerUpdaterWithSettings(settings)

	return func(players map[int]Glicko2Player, matches []Glicko2MatchByID) (map[int]Glicko2Player, error) {

		updatedPlayers := make(map[int]Glicko2Player)

		newMatchLists := make(map[int][]Glicko2MatchForPlayer)

		for _, match := range matches {
			//match = match
			newMatchLists[match.Player1ID] = append(newMatchLists[match.Player1ID], Glicko2MatchForPlayer{
				Opponent: players[match.Player2ID],
				Result:   match.Result,
			},
			)

			newMatchLists[match.Player2ID] = append(newMatchLists[match.Player2ID], Glicko2MatchForPlayer{
				Opponent: players[match.Player1ID],
				Result:   1 - match.Result,
			},
			)

		}

		for playerID, player := range players {

			var playerMatchList []Glicko2MatchForPlayer
			if filledMatchList, ok := newMatchLists[playerID]; ok {
				playerMatchList = filledMatchList
			}
			updatedPlayer, err := newPlayerUpdater(player, playerMatchList)

			if err != nil {
				return nil, err
			}

			updatedPlayers[playerID] = updatedPlayer

		}

		return updatedPlayers, nil
	}
}

func DefaultPeriodCalculator() func(players map[int]Glicko2Player, matches []Glicko2MatchByID) (map[int]Glicko2Player, error) {
	return PeriodCalculatorWithSettings(glicko2DefaultSettings)
}
