package main

import "fmt"

// Location ...Might try to add meta data if it seems worthwhile
type Location struct {
	X int	`json:"x"`
	Y int	`json:"y"`
}

var directions = []Location{
	{0, 1},
	{1, 0},
	{0, -1},
	{-1, 0},
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
	allySpaces []Location,
	enemySpaces []Location,
) {

	var currentTile = terrainMap[playerLocation.X][playerLocation.Y]

	for _, direction := range directions {
		newX, newY := playerLocation.X+direction.X, playerLocation.Y+direction.Y
		var neighbor = Location{
			newX,
			newY,
		}

		if neighbor.X >= 0 && neighbor.X <= len(terrainMap)-1 && neighbor.Y >= 0 && neighbor.Y <= len(terrainMap[0])-1 {
			neighborTile := terrainMap[neighbor.X][neighbor.Y]

			// in reality the following items still need to be taken into account:
			// 1. a character can jump down further than they can jump up, ?(jump vs. jump - 1)?	#needsTesting
			// 2. the cost for jumping up is higher than jumping down (2 vs 1)						#needsTesting
			// 3. certain types of terrain cost more to move across (probably only matters for water)
			// 4. jumping across gaps or over enemies (long term goal)
			// 5. occupied (thus not movable) spaces for allies (passable) and enemies (blocked)	#needsTesting
			cost := 1
			fmt.Println(neighborTile)
			if neighborTile.Height-currentTile.Height > jump/2 {
				cost = 2
			} else if neighborTile.Terrain == "deep_water" {
				cost = move
			}
			//if math.Abs(float64(neighborTile.Height-currentTile.Height)) < float64(jump) {
			if (neighborTile.Height-currentTile.Height < jump ||
				currentTile.Height-neighborTile.Height <= jump) &&
				!ContainsPoint(enemySpaces, neighbor) {

				// recursively look for other movable locations if player still has moves left
				if move > cost {
					getNeighbors(
						move-cost,
						jump,
						neighbor,
						terrainMap,
						movableLocations,
						allySpaces,
						enemySpaces,
					)
				}

				if !ContainsPoint(*movableLocations, neighbor) && !ContainsPoint(allySpaces, neighbor) {
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
	allySpaces []string,
	enemySpaces []string,
) []Location {
	var movableLocations = []Location{playerLocation}
	allySpacesClean := ConvertStringToCoords(allySpaces)
	enemySpacesClean := ConvertStringToCoords(enemySpaces)

	getNeighbors(move, jump, playerLocation, terrainMap, &movableLocations, allySpacesClean, enemySpacesClean)

	return movableLocations
}

func getNextStep(
	move int,
	jump int,
	playerLocation Location,
	newLocation Location,
	terrainMap [][]MapTile,
	enemySpaces []Location,
	pathSoFar []Location,
) []Location {
	// TODO: a lot of the conditionals below are repeats of algorithm above. extract into function if applicable
	var currentTile = terrainMap[playerLocation.X][playerLocation.Y]

	for _, direction := range directions {
		newX, newY := playerLocation.X+direction.X, playerLocation.Y+direction.Y
		var neighbor = Location{
			newX,
			newY,
		}

		if neighbor.X >= 0 && neighbor.X <= len(terrainMap)-1 && neighbor.Y >= 0 && neighbor.Y <= len(terrainMap[0])-1 {
			neighborTile := terrainMap[neighbor.X][neighbor.Y]

			cost := 1
			if neighborTile.Height-currentTile.Height > jump/2 {
				cost = 2
			} else if neighborTile.Terrain == "deep_water" {
				cost = move
			}

			if (neighborTile.Height-currentTile.Height < jump ||
				currentTile.Height-neighborTile.Height <= jump) &&
				!ContainsPoint(enemySpaces, neighbor) {
				newPath := append(append([]Location{}, pathSoFar...), neighbor)
				// check if it is the desired location
				if neighbor.X == newLocation.X && neighbor.Y == newLocation.Y {
					return newPath
				} else if move > cost {

					return getNextStep(
						move,
						jump,
						neighbor,
						newLocation,
						terrainMap,
						enemySpaces,
						newPath,
					)
				}
			}
		}
	}
}

func GetPath(
	move int,
	jump int,
	playerLocation Location,
	newLocation Location,
	terrainMap [][]MapTile,
	enemySpaces []string,
) []Location {
	var bestPath []Location
	// keeping it simple as the cost of path doesn't affect anything and move distance should be 3-5 realistically
	// essentially breadth first search
	// 1. run function on all neighbor tiles that checks if they are accessible and then starts a list
	// 2. recursively run this function on neighbor tiles still in frontier until one of the paths reaches the endpoint
	// 3. return list of tiles for successful path as json
	bestPath = getNextStep(
		move,
		jump,
		playerLocation,
		newLocation,
		terrainMap,
		ConvertStringToCoords(enemySpaces),
		[]Location{},
	)
	return bestPath
}
