package tactics

import (
	"fmt"
	"strconv"
	"strings"
)

func runPaths(start []int, end []int) int {
	if start[0] == end[0] || start[1] == end[1] {
		return 0
	} else {
		return 1 + runPaths([]int{start[0], start[1] + 1}, end) + runPaths([]int{start[0] + 1, start[1]}, end)
	}
}

func getPaths(pointSet string) int {
	var startPoint []int
	var endPoint []int

	var uglyPoints []string = strings.Split(strings.Trim(pointSet, "()"), ")(")

	stringStartPoint := strings.Split(uglyPoints[0], " ")
	stringEndPoint := strings.Split(uglyPoints[1], " ")

	for _, i := range stringStartPoint {
		value, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		startPoint = append(startPoint, value)
	}
	for _, i := range stringEndPoint {
		value, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		endPoint = append(endPoint, value)
	}

	fmt.Println(startPoint)
	fmt.Println(endPoint)

	fmt.Printf("The number of paths is: %d\n", runPaths(startPoint, endPoint)+1)

	return runPaths(startPoint, endPoint)
}

func main() {
	getPaths("(1 1)(1 1)")
	getPaths("(0 0)(1 1)")
	getPaths("(3 3)(5 5)")
	getPaths("(2 5)(3 7)")
	getPaths("(1 3)(5 5)")
}
