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


	// TODO: find a way to do this dynamically, so we can have a dynamic number of allies and enemies
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
		"SELECT id, name, maxhp, maxmp, move, jump, speed FROM enemy WHERE id IN ($1, $2, $3)",
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

	// create battle id to store all pertinent information
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
	var characterTimeInfo []TimeInfo
	for i := 0; i < len(battleConfiguration.Allies); i++ {
		currAlly := allies[i]
		allyRef := battleId + ":allies:" + strconv.Itoa(currAlly.Id)
		allyTimeGauge := rand.Intn(800)
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
			allyTimeGauge,
		).Err
		if err != nil {
			fmt.Println("adding allies error")
			log.Fatalln(err)
		}
		// Adds location to a list of occupied spaces as "x:y"
		err = conn.Cmd(
			"SADD",
			battleId + ":allySpaces",
			strconv.Itoa(battleConfiguration.AllyLocations[i].X) + ":" + strconv.Itoa(battleConfiguration.AllyLocations[i].Y),
		).Err
		if err != nil {
			log.Fatalln(err)
		}


		characterTimeInfo = append(
			characterTimeInfo,
			TimeInfo{
				currAlly.Id,
				currAlly.Speed,
				allyTimeGauge,
				false,
			})
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
		enemyTimeGauge := rand.Intn(800)

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
			"timeGauge",
			enemyTimeGauge,
		).Err
		if err != nil {
			fmt.Println("Redis: adding enemies error", currEnemy)
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

		characterTimeInfo = append(
			characterTimeInfo,
			TimeInfo{
				i,
				currEnemy.Speed,
				enemyTimeGauge,
				true,
			})
	}

	turnOrder := GetTurnOrder(characterTimeInfo)

	for _, characterTurn := range turnOrder {
		var characterRef string
		if characterTurn.IsEnemy {
			characterRef = battleId + ":enemies:" + strconv.Itoa(characterTurn.Id)
		} else {
			characterRef = battleId + ":allies:" + strconv.Itoa(characterTurn.Id)
		}
		err = conn.Cmd("HSET", characterRef, "timeGauge", characterTurn.TimeGauge).Err
		if err != nil {
			fmt.Println("Redis: updating timeGauges error")
			log.Fatalln(err)
		}
	}

	// TODO: check if any enemies have their turn and handle their moves

	// storing information on character who is taking their turn
	err = conn.Cmd(
		"HMSET",
		battleId + ":activePlayer",
		"id",
		turnOrder[0].Id,
		"isEnemy",
		turnOrder[0].IsEnemy,
		"actionsTaken",
		0,
	).Err

	json.NewEncoder(w).Encode(struct {
		BattleId 			string	 			`json:"battleId"`
		TurnOrder			[]TimeInfo			`json:"turnOrder"`
	}{
		battleId[7:],
		turnOrder,
	})

}
