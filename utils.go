package main

import (
	"strings"
	"strconv"
)

func ConvertStringToCoords(stringArray []string) []Location {
	output := make([]Location, len(stringArray))
	for i, value := range stringArray {
		currPoint := strings.Split(value, ":")
		// TODO: error handling below...
		xPoint, _ := strconv.Atoi(currPoint[0])
		yPoint, _ := strconv.Atoi(currPoint[1])
		output[i] = Location{xPoint, yPoint}
	}
	return output
}
