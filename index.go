package main

import (
	_ "github.com/lib/pq"
	"github.com/husobee/vestigo"
	"net/http"
	"log"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix.v2/pool"
	)

type AppHandler struct {

}

func main() {

	dbInfo := fmt.Sprintf(
		"user='vick michael' password='' dbname='vick michael' sslmode=disable",
	)
	db, err := sqlx.Connect("postgres", dbInfo)
	if err != nil {
		fmt.Println("check your Postgres connection")
		panic(err)
	}

	redisPool, redisErr := pool.New("tcp", "localhost:6379", 10)
	if redisErr != nil {
		fmt.Println("check your Redis connection")
		panic(err)
	}

	// Controller declarations (Postgres)
	mapController := MapController{db}
	characterController := CharacterController{redisPool, db}
	enemyController := EnemyController{db}

	// Controller declarations (Redis)
	battleManager := BattleManagementController{redisPool, db}

	// Router declaration
	router := vestigo.NewRouter()

	// Setting up router global  CORS policy
	// These policy guidelines are overriddable at a per resource level shown below
	router.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:      []string{"*"},
		AllowCredentials: true,
		// ExposeHeaders:    []string{"X-Header", "X-Y-Header"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers", "Content-Type"},
	})

	// Maps routes
	router.Get("/maps", mapController.getMaps)
	router.Get("/maps/:id", mapController.getMap)
	router.Post("/maps/create", mapController.createMap)

	// Characters routes
	router.Get("/characters", characterController.getCharacters)
	router.Get("/characters/:id", characterController.getCharacter)
	router.Post("/characters/create", characterController.createCharacter)

	// Enemies routes
	router.Get("/enemies", enemyController.getEnemies)
	router.Get("/enemies/:id", enemyController.getEnemy)
	router.Post("/enemies/create", enemyController.createEnemy)

	// Characters' movement routes
	// TODO: transfer these routes and the columns in postgres to battle management routes and Redis data
	router.Get("/battle/:battleId/movableLocations/:id", characterController.getMovableSpaces)
	router.Post("/characters/:battleId/move/:id", characterController.move)

	// Battle managing routes
	router.Post("/battle/initialize", battleManager.initializeBattle)

	log.Fatal(http.ListenAndServe(":1234", router))
}
