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
	"math/rand"
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
		"SELECT * FROM character WHERE id IN ($1, $2, $3)",
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
		"SELECT name, maxhp, maxmp, move, jump FROM enemy WHERE id IN ($1, $2, $3)",
		battleConfiguration.Enemies[0],
		battleConfiguration.Enemies[1],
		battleConfiguration.Enemies[2],
	)
	if err != nil {
		log.Println(err)
	}

	conn, err := bmController.redisPool.Get()
	if err != nil {
		fmt.Println("couldn't get Redis pool connection")
		log.Fatalln(err)
		return
	}
	defer bmController.redisPool.Put(conn)

	// TODO: make a better method for creating a random hash
	var battleId string
	exists := 1
	for exists == 1 {
		battleId = "battle:" + RandomString(32)[:30]
		exists, err = conn.Cmd("EXISTS", battleId + ":mapId").Int()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	err = conn.Cmd("SET", battleId + ":mapId", battleConfiguration.MapId).Err

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
			"move",
			currAlly.Move,
			"jump",
			currAlly.Jump,
			"HP",
			currAlly.MaxHP,
			"maxHP",
			currAlly.MaxHP,
			"MP",
			currAlly.MaxMP,
			"maxMP",
			currAlly.MaxMP,
			"x",
			battleConfiguration.AllyLocations[i].X,
			"y",
			battleConfiguration.AllyLocations[i].Y,
			"timeGauge",
			currAlly.Speed * rand.Intn(5),
		).Err
		if err != nil {
			fmt.Println("adding allies error")
			log.Fatalln(err)
		}
		// TODO: Adds location to a list of occupied spaces as "x:y"
		err = conn.Cmd(
			"SADD",
			battleId + ":allySpaces",
			strconv.Itoa(battleConfiguration.AllyLocations[i].X) + ":" + strconv.Itoa(battleConfiguration.AllyLocations[i].Y),
		).Err
		if err != nil {
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

		enemyRef := battleId + ":enemies:" + strconv.Itoa(i)
		err = conn.Cmd(
			"HMSET",
			enemyRef,
			"id",
			currEnemy.Id,
			"name",
			currEnemy.Name,
			"move",
			currEnemy.Move,
			"jump",
			currEnemy.Jump,
			"HP",
			currEnemy.MaxHP,
			"maxHP",
			currEnemy.MaxHP,
			"MP",
			currEnemy.MaxMP,
			"maxMP",
			currEnemy.MaxMP,
			"x",
			battleConfiguration.EnemyLocations[i].X,
			"y",
			battleConfiguration.EnemyLocations[i].Y,
		).Err
		if err != nil {
			fmt.Println("adding enemies error", currEnemy)
			log.Fatalln(err)
		}

		err = conn.Cmd(
			"SADD",
			battleId + ":enemySpaces",
			strconv.Itoa(battleConfiguration.EnemyLocations[i].X) + ":" + strconv.Itoa(battleConfiguration.EnemyLocations[i].Y),
		).Err
		if err != nil {
			log.Fatalln(err)
		}
	}

	json.NewEncoder(w).Encode(struct {
		BattleId 	string	 `json:"battleId"`
	}{
		battleId[7:],
	})
	//enemy0, err := conn.Cmd("HGETALL", battleId + ":enemies:0").Map()
	//if err != nil {
	//	fmt.Println("Redis GET command failed")
	//	log.Fatalln(err)
	//}
	//fmt.Println(enemy0)
	//
	//enemy1, err := conn.Cmd("HGETALL", battleId + ":enemies:1").Map()
	//if err != nil {
	//	fmt.Println("Redis GET command failed")
	//	log.Fatalln(err)
	//}
	//fmt.Println(enemy1)
	//
	//enemy2, err := conn.Cmd("HGETALL", battleId + ":enemies:2").Map()
	//if err != nil {
	//	fmt.Println("Redis GET command failed")
	//	log.Fatalln(err)
	//}
	//fmt.Println(enemy2)
}
