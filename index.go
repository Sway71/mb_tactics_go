package main

import (
	_ "github.com/lib/pq"
	"github.com/husobee/vestigo"
	"net/http"
	"log"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type AppHandler struct {

}

func main() {

	dbInfo := fmt.Sprintf(
		"user='vick michael' password='' dbname='vick michael' sslmode=disable",
	)
	db, err := sqlx.Connect("postgres", dbInfo)
	if err != nil {
		panic(err)
	}

	mapController := MapController{db}
	characterController := CharacterController{db}

	router := vestigo.NewRouter()

	// manage Maps
	router.Get("/maps", mapController.getMaps)
	router.Get("/maps/:id", mapController.getMap)
	router.Post("/maps/create", mapController.createMap)

	// manage Characters
	router.Get("/characters", characterController.getCharacters)
	router.Get("/characters/:id", characterController.getCharacter)
	router.Post("/characters/create", characterController.createCharacter)

	// Characters' movement
	router.Get("/characters/:id/movableLocations", characterController.getMovableSpaces)
	router.Post("/characters/:id/move", characterController.move)

	router.Post("/welcome/:name", PostWelcomeHandler)

	log.Fatal(http.ListenAndServe(":1234", router))
}

// PostWelcomeHandler - Is an Implementation of http.HandlerFunc
func PostWelcomeHandler(w http.ResponseWriter, r *http.Request) {
	name := vestigo.Param(r, "name") // url params live in the request
	w.WriteHeader(200)
	w.Write([]byte("welcome " + name +"!"))
}

//func GetAvailableSpaces(w http.ResponseWriter, r *http.Request) {
//	newPlayer := Character{
//		5,
//		"Sam",
//		3,
//		1,
//		2,
//		2,
//	}
//
//	var newPlayerLocation = Location{newPlayer.X, newPlayer.Y}
//
//	var terrainHeightMap = [][]int{
//		{0, 0, 1, 1, 1},
//		{0, 0, 1, 2, 2},
//		{0, 0, 2, 3, 1},
//		{0, 0, 4, 2, 2},
//		{0, 0, 4, 3, 2},
//	}
//	var ListOfSpaces = getMovableSpaces(3, newPlayerLocation, terrainHeightMap)
//
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(200)
//	json.NewEncoder(w).Encode(ListOfSpaces)
//	// w.Write([]byte(listOfSpaces))
//}
