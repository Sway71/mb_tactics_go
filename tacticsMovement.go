package main

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

		if neighbor.X >= 0 && neighbor.X <= len(terrainMap)-1 && neighbor.Y >= 0 && neighbor.Y <= len(terrainMap[0])-1 {
			neighborTile := terrainMap[neighbor.X][neighbor.Y]

			// in reality the following items still need to be taken into account:
			// 1. a character can jump down further than they can jump up, ?(jump vs. jump - 1)?	#needsTesting
			// 2. the cost for jumping up is higher than jumping down (2 vs 1)						#needsTesting
			// 3. certain types of terrain cost more to move across (probably only matters for water)
			// 4. jumping across gaps or over enemies (long term goal)
			cost := 1
			if neighborTile.Height-currentTile.Height > jump/2 {
				cost = 2
			}
			//if math.Abs(float64(neighborTile.Height-currentTile.Height)) < float64(jump) {
			if neighborTile.Height-currentTile.Height < jump || currentTile.Height-neighborTile.Height <= jump {

				// recursively look for other movable locations if player still has moves left
				if move > cost {
					getNeighbors(move-cost, jump, neighbor, terrainMap, movableLocations)
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

	getNeighbors(move, jump, playerLocation, terrainMap, &movableLocations)

	return movableLocations
}

func GetPath(
	move int,
	jump int,
	playerLocation Location,
	terrainMap [][]MapTile,
) []Location {
	var bestPath []Location
	return bestPath
}
