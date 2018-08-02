package main

import (
		"sort"
	"fmt"
)

type TimeInfo struct {
	Id        int
	Speed     int
	TimeGauge int
	Enemy     bool
}

func GetTurnOrder(charactersTimeInfo []TimeInfo) ([]TimeInfo) {
	numTurnsPassed := 1000
	sort.Slice(charactersTimeInfo, func(i, j int) bool {
		iTurnsLeft := (1000 - charactersTimeInfo[i].TimeGauge) / charactersTimeInfo[i].Speed
		jTurnsLeft := (1000 - charactersTimeInfo[j].TimeGauge) / charactersTimeInfo[j].Speed
		fmt.Println(iTurnsLeft)
		if iTurnsLeft < numTurnsPassed {
			numTurnsPassed = iTurnsLeft
		}
		return jTurnsLeft > iTurnsLeft
	})

	for index, character := range charactersTimeInfo {
		if index == 0 {
			fmt.Println(1000 - charactersTimeInfo[index].TimeGauge - (character.Speed * numTurnsPassed))
			charactersTimeInfo[index].TimeGauge = 1000
		} else {
			charactersTimeInfo[index].TimeGauge = character.TimeGauge + (numTurnsPassed * character.Speed)
		}
	}

	return charactersTimeInfo
}
