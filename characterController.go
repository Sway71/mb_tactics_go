package main

import (
	"net/http"
	"github.com/husobee/vestigo"
	"encoding/json"
	"log"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"math/rand"
	"fmt"
	"github.com/mediocregopher/radix.v2/pool"
	"strconv"
)

type Character struct {
	Id 			int				`json:"id"`
	Name 		string			`json:"name"`
	Level		int				`json:"level"`
	Exp			int				`json:"exp"`
	HP			int				`json:"hp"`
	MaxHP		int				`json:"maxHp"`
	MP			int				`json:"mp"`
	MaxMP		int				`json:"maxMp"`
	Strength	int				`json:"strength"`
	Speed		int				`json:"speed"`
	Move		int				`json:"move"`
	Jump 		int				`json:"jump"`
}

type CharacterController struct {
	redisPool		*pool.Pool
	DB				*sqlx.DB
}

func (c *CharacterController) getCharacters(w http.ResponseWriter, r *http.Request) {
	var characters []Character
	err := c.DB.Select(&characters, "SELECT * FROM character ORDER BY id")
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(characters)
}

func (c *CharacterController) getCharacter(w http.ResponseWriter, r *http.Request) {
	var character Character
	id := vestigo.Param(r, "id")

	err := c.DB.Get(&character, "SELECT * FROM character WHERE id=$1", id)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(character)
}

func (c *CharacterController) createCharacter(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var character Character
	err = json.Unmarshal(b, &character)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var lastInsertId int;
	characterInsert := `
		INSERT INTO character (
			name,
			level,
			exp,
			hp,
			mp,
			move,
			jump,
			speed,
			x,
			y
		  ) VALUES (
		  	$1,
		  	$2,
		  	$3,
		  	$4,
		  	$5,
		  	$6,
		  	$7,
		  	$8,
		  	$9,
		  	$10,
			$11,
			$12,
			$13,
			$14
		  ) RETURNING id
	`
	hp := rand.Intn(10) + 25
	mp := rand.Intn(5) + 8
	err = c.DB.QueryRow(
		characterInsert,
		character.Name,
		0,
		0,
		hp,
		mp,
		3,
		3,
		5,
		0,
		0,
	).Scan(&lastInsertId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(struct {
		Id		int		`json:"id"`
	}{
		lastInsertId,
	})
}

func (c *CharacterController) getMovableSpaces(w http.ResponseWriter, r *http.Request) {
	id := vestigo.Param(r, "id")
	battleId := vestigo.Param(r, "battleId")

	conn, err := c.redisPool.Get()
	if err != nil {
		fmt.Println("couldn't get Redis pool connection")
		log.Fatalln(err)
		return
	}
	defer c.redisPool.Put(conn)

	battlefieldId, err := conn.Cmd("GET", "battle:" + battleId + ":mapId").Str()

	var battlefield Map
	err = c.DB.Get(&battlefield, "SELECT * FROM battlefield WHERE id=$1", battlefieldId)
	if err != nil {
		log.Fatalln(err)
	}

	var battlefieldLayout [][]MapTile
	json.Unmarshal(battlefield.MapData, &battlefieldLayout)


	ally, err := conn.Cmd(
		"HMGET",
		"battle:" + battleId + ":allies:" + id,
		"x",
		"y",
		"move",
		"jump",
	).List()
	if err != nil {
		log.Fatalln(err)
	}

	allyX, _ := strconv.Atoi(ally[0])
	allyY, _ := strconv.Atoi(ally[1])
	allyMove, _ := strconv.Atoi(ally[2])
	allyJump, _ := strconv.Atoi(ally[3])

	json.NewEncoder(w).Encode(GetMovableSpaces(
		allyMove,
		allyJump,
		Location{allyX, allyY},
		battlefieldLayout,
	))
}

func (c *CharacterController) move(w http.ResponseWriter, r *http.Request) {
	id := vestigo.Param(r, "id")
	battleId := vestigo.Param(r, "battleId")

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var location Location
	err = json.Unmarshal(b, &location)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	conn, err := c.redisPool.Get()
	if err != nil {
		fmt.Println("couldn't get Redis pool connection")
		log.Fatalln(err)
		return
	}
	defer c.redisPool.Put(conn)

	battlefieldId, err := conn.Cmd("GET", "battle:" + battleId + ":mapId").Str()

	var battlefield Map
	err = c.DB.Get(&battlefield, "SELECT * FROM battlefield WHERE id=$1", battlefieldId)
	if err != nil {
		log.Fatalln(err)
	}

	var battlefieldLayout [][]MapTile
	json.Unmarshal(battlefield.MapData, &battlefieldLayout)


	ally, err := conn.Cmd(
		"HMGET",
		"battle:" + battleId + ":allies:" + id,
		"x",
		"y",
		"move",
		"jump",
	).List()
	if err != nil {
		log.Fatalln(err)
	}

	allyX, _ := strconv.Atoi(ally[0])
	allyY, _ := strconv.Atoi(ally[1])
	allyMove, _ := strconv.Atoi(ally[2])
	allyJump, _ := strconv.Atoi(ally[3])

	validMove := ContainsPoint(GetMovableSpaces(
		allyMove,
		allyJump,
		Location{allyX, allyY},
		battlefieldLayout,
	), location)
	fmt.Println(validMove)
	//
	//if (validMove) {
	//	characterMove := "UPDATE character SET x = $1, y = $2 WHERE id = $3 RETURNING x, y"
	//	result := c.DB.MustExec(characterMove, location.X, location.Y, id)
	//	rowsAffected, _ := result.RowsAffected()
	//
	//	wasSuccessful := rowsAffected == 1
	//	json.NewEncoder(w).Encode(struct {
	//		Success bool `json:"success"`
	//	}{
	//		wasSuccessful,
	//	})
	//} else {
	//	json.NewEncoder(w).Encode(struct {
	//		Success bool `json:"success"`
	//	}{
	//		false,
	//	})
	//}
}
//
//func (c *CharacterController) attack(w http.ResponseWriter, r *http.Request) {
//
//}
//
//func (c *CharacterController) special(w http.ResponseWriter, r *http.Request) {
//
//}