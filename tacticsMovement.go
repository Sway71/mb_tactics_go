package main

import (
	"math"
)

// Location ...Might try to add meta data if it seems worthwhile
type Location struct {
	X int	`json:"x"`
	Y int	`json:"y"`
}

func ContainsPoint(pointsSlice []Location, pointToCheck Location) bool {
	result := false
	for _, point := range pointsSlice {
		if point.X == pointToCheck.X && point.Y == pointToCheck.Y {
			result = true
		}
	}
	return result
}

func getNeighbors(
	move int,
	playerLocation Location,
	terrainMap [][]int,
	movableLocations *[]Location,
) {
	var directions = []Location{
		{0, 1},
		{1, 0},
		{0, -1},
		{-1, 0},
	}

	var currentHeight = terrainMap[playerLocation.X][playerLocation.Y]

	for _, direction := range directions {
		newX, newY := playerLocation.X+direction.X, playerLocation.Y+direction.Y
		var neighbor = Location{
			newX,
			newY,
		}

		if neighbor.X >= 0 && neighbor.X <= 4 && neighbor.Y >= 0 && neighbor.Y <= 4 {
			neighborHeight := terrainMap[neighbor.X][neighbor.Y]

			if math.Abs(float64(neighborHeight-currentHeight)) < 2 {

				// recursively look for other movable locations if player still has moves left
				if move > 1 {
					getNeighbors(move-1, neighbor, terrainMap, movableLocations)
				}

				if !ContainsPoint(*movableLocations, neighbor) {
					*movableLocations = append(*movableLocations, neighbor)
				}
			}
		}

	}
}

func GetMovableSpaces(
	move int,
	playerLocation Location,
	// terrainMap [][]int,
) []Location {
	var movableLocations = []Location{playerLocation}

	var terrainMap = [][]int{
		{0, 0, 1, 1, 2, 3, 5},
		{0, 0, 1, 2, 2, 5, 7},
		{0, 0, 2, 3, 1, 3, 9},
		{0, 0, 4, 2, 2, 3, 4},
		{0, 0, 4, 3, 2, 3, 4},
		{0, 0, 4, 3, 2, 2, 1},
		{0, 0, 4, 3, 2, 1, 1},
	}

	getNeighbors(move, playerLocation, terrainMap, &movableLocations)

	return movableLocations
	//fmt.Printf("number of movable locations: %d\n", len(movableLocations))
	//fmt.Printf("Set of movable locations: \n\t %v", movableLocations)
}

//func main() {
//	var newPlayer = Character{
//		"Sam",
//		3,
//		1,
//		Location{2, 2},
//	}
//
//	var terrainHeightMap = [][]int{
//		{0, 0, 1, 1, 1},
//		{0, 0, 1, 2, 2},
//		{0, 0, 2, 3, 1},
//		{0, 0, 4, 2, 2},
//		{0, 0, 4, 3, 2},
//	}
//	getMovableSpaces(3, newPlayer.location, terrainHeightMap)
//}
