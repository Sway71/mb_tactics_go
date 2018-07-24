package main

import (
	"github.com/mediocregopher/radix.v2/pool"
	"log"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"github.com/jmoiron/sqlx"
)

type BattleManagementController struct {
	redisPool		*pool.Pool
	DB				*sqlx.DB
}

type BattleConfiguration struct {
	MapId					int					`json:"mapId"`
	Allies					[]int				`json:"allies"`
	AllyLocations			[]Location			`json:"allyLocations"`
	Enemies					[]int				`json:"enemies"`
	EnemyLocations			[]Location			`json:"enemyLocations"`
}


func (bmController *BattleManagementController) initializeBattle(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var battleConfiguration BattleConfiguration
	err = json.Unmarshal(b, &battleConfiguration)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}



	var allies []Character
	err = bmController.DB.Select(
		&allies,
		"SELECT id, name, maxhp, maxmp FROM character WHERE id IN ($1, $2, $3)",
		battleConfiguration.Allies[0],
		battleConfiguration.Allies[1],
		battleConfiguration.Allies[2],
	)
	if err != nil {
		log.Println(err)
	}

	var enemies []Enemy
	err = bmController.DB.Select(
		&enemies,
		"SELECT id, name, maxhp, maxmp FROM enemy WHERE id IN ($1, $2, $3)",
		battleConfiguration.Enemies[0],
		battleConfiguration.Enemies[1],
		battleConfiguration.Enemies[2],
	)
	fmt.Println(enemies)
	if err != nil {
		log.Println(err)
	}

	// TODO: make a better method for creating a random hash
	battleId := "battle:" + RandomString(20)

	conn, err := bmController.redisPool.Get()
	if err != nil {
		fmt.Println("couldn't get Redis pool connection")
		log.Fatalln(err)
	}
	defer bmController.redisPool.Put(conn)

	// Add allies to battle in Redis
	for i := 0; i < len(battleConfiguration.Allies); i++ {
		currAlly := allies[i]
		allyRef := battleId + ":allies:" + strconv.Itoa(currAlly.Id)
		err = conn.Cmd(
			"HMSET",
			allyRef,
			"id",
			currAlly.Id,
			"name",
			currAlly.Name,
			"HP",
			currAlly.MaxHP,
			"maxHP",
			currAlly.MaxHP,
			"MP",
			currAlly.MaxMP,
			"maxMP",
			currAlly.MaxMP,
		).Err
		if err != nil {
			fmt.Println("adding allies error")
			log.Fatalln(err)
		}
	}

	// Add enemies to battle in Redis
	for i := 0; i < len(battleConfiguration.Enemies); i++ {
		var currEnemy Enemy
		for _, enemyData := range enemies {
			if enemyData.Id == battleConfiguration.Enemies[i] {
				currEnemy = enemyData
			}
		}
		// currEnemy = enemies[i]
		enemyRef := battleId + ":enemies:" + strconv.Itoa(i)
		err = conn.Cmd(
			"HMSET",
			enemyRef,
			"id",
			currEnemy.Id,
			"name",
			currEnemy.Name,
			"HP",
			currEnemy.MaxHP,
			"maxHP",
			currEnemy.MaxHP,
			"MP",
			currEnemy.MaxMP,
			"maxMP",
			currEnemy.MaxMP,
		).Err
		if err != nil {
			fmt.Println("adding enemies error", currEnemy)
			log.Fatalln(err)
		}
	}

	enemy0, err := conn.Cmd("HGETALL", battleId + ":enemies:0").Map()
	if err != nil {
		fmt.Println("Redis GET command failed")
		log.Fatalln(err)
	}
	fmt.Println(enemy0)

	enemy1, err := conn.Cmd("HGETALL", battleId + ":enemies:1").Map()
	if err != nil {
		fmt.Println("Redis GET command failed")
		log.Fatalln(err)
	}
	fmt.Println(enemy1)

	enemy2, err := conn.Cmd("HGETALL", battleId + ":enemies:2").Map()
	if err != nil {
		fmt.Println("Redis GET command failed")
		log.Fatalln(err)
	}
	fmt.Println(enemy2)
}
