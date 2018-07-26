package main

import (
	"math"
	"fmt"
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
	jump int,
	playerLocation Location,
	terrainMap [][]MapTile,
	movableLocations *[]Location,
) {
	var directions = []Location{
		{0, 1},
		{1, 0},
		{0, -1},
		{-1, 0},
	}

	var currentTile = terrainMap[playerLocation.X][playerLocation.Y]

	for _, direction := range directions {
		newX, newY := playerLocation.X+direction.X, playerLocation.Y+direction.Y
		var neighbor = Location{
			newX,
			newY,
		}
		fmt.Println()
		if neighbor.X >= 0 && neighbor.X <= len(terrainMap)-1 && neighbor.Y >= 0 && neighbor.Y <= len(terrainMap[0])-1 {
			neighborTile := terrainMap[neighbor.X][neighbor.Y]

			if math.Abs(float64(neighborTile.Height-currentTile.Height)) < float64(jump) {

				// recursively look for other movable locations if player still has moves left
				if move > 1 {
					getNeighbors(move-1, jump, neighbor, terrainMap, movableLocations)
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
	jump int,
	playerLocation Location,
	terrainMap [][]MapTile,
) []Location {
	var movableLocations = []Location{playerLocation}

	//var terrainMap = [][]int{
	//	{0, 0, 1, 1, 2, 3, 5},
	//	{0, 0, 1, 2, 2, 5, 7},
	//	{0, 0, 2, 3, 1, 3, 9},
	//	{0, 0, 4, 2, 2, 3, 4},
	//	{0, 0, 4, 3, 2, 3, 4},
	//	{0, 0, 4, 3, 2, 2, 1},
	//	{0, 0, 4, 3, 2, 1, 1},
	//}

	getNeighbors(move, jump, playerLocation, terrainMap, &movableLocations)

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
