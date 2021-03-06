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
	EquipmentId	int				`json:"equipmentId" db:"equipment_id"`
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

	allySpaces, err := conn.Cmd("SMEMBERS", "battle:" + battleId + ":allySpaces").List()
	enemySpaces, err := conn.Cmd("SMEMBERS", "battle:" + battleId + ":enemySpaces").List()

	json.NewEncoder(w).Encode(GetMovableSpaces(
		allyMove,
		allyJump,
		Location{allyX, allyY},
		battlefieldLayout,
		allySpaces,
		enemySpaces,
	))
}

// TODO: move to battle management controller or its own movement controller.
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

	// validate that it is this character's turn
	actionsTaken, err := conn.Cmd(
		"HGET",
		"battle:" + battleId + ":activePlayer",
		"actionsTaken",
	).Int()
	if actionsTaken == 1 || actionsTaken == 3 || err != nil {
		// not the best response, but it should work for now
		http.Error(w, err.Error(), 401)
		return
	} else {
		// still unsure, but planning on having move count as a 1 and attack count as 2
		// thus moved == 1, attacked == 2, attacked and moved == 3, turn just started == 0
		err := conn.Cmd(
			"HINCRBY",
			"battle:" + battleId + ":activePlayer",
			"actionsTaken",
			1,
		).Err
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

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


	allySpaces, err := conn.Cmd("SMEMBERS", "battle:" + battleId + ":allySpaces").List()
	enemySpaces, err := conn.Cmd("SMEMBERS", "battle:" + battleId + ":enemySpaces").List()

	validMove := ContainsPoint(GetMovableSpaces(
		allyMove,
		allyJump,
		Location{allyX, allyY},
		battlefieldLayout,
		allySpaces,
		enemySpaces,
	), location)

	if validMove {

		// get path to destination, update location and occupied tiles in Redis, then send back the path
		pathToDestination := GetPath(
			allyMove,
			allyJump,
			Location{allyX, allyY},
			location,
			battlefieldLayout,
			enemySpaces,
		)

		err = conn.Cmd("SREM", "battle:" + battleId + ":allySpaces", ally[0] + ":" + ally[1]).Err
		err = conn.Cmd("SADD", "battle:" + battleId + ":allySpaces", strconv.Itoa(location.X) + ":" + strconv.Itoa(location.Y)).Err
		err = conn.Cmd("HMSET", "battle:" + battleId + ":allies:" + id, "x", location.X, "y", location.Y).Err

		//
		//err = conn.Cmd("HINCRBY", "battle:" + battleId + ":allies:" + id, ).Err
		if err != nil {
			// TODO: decide on proper error to throw
			fmt.Println("error trying to update new ally location in Redis")
		}
		//allyValues, _ := conn.Cmd("HGETALL", "battle:" + battleId + ":allies:" + id).List()
		//fmt.Println(allyValues)

		json.NewEncoder(w).Encode(struct {
			Success 			bool			`json:"success"`
			PathToDestination	[]Location		`json:"pathToDestination"`
		}{
			validMove,
			pathToDestination,
		})
	} else {
		json.NewEncoder(w).Encode(struct {
			Success 			bool		 	`json:"success"`
		}{
			false,
		})
	}
}

// TODO: move to battle management controller or its own attack controller.
func (c *CharacterController) attack(w http.ResponseWriter, r *http.Request) {
	// id := vestigo.Param(r, "id")
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

	// validate that it is this character's turn

	actionsTaken, err := conn.Cmd(
		"HGET",
		"battle:" + battleId + ":activePlayer",
		"actionsTaken",
	).Int()
	if actionsTaken == 2 || actionsTaken == 3 || err != nil {
		// not the best response, but it should work for now
		http.Error(w, err.Error(), 401)
		return
	} else {
		// still unsure, but planning on having move count as a 1 and attack count as 2
		// thus moved == 1, attacked == 2, attacked and moved == 3, turn just started == 0
		err := conn.Cmd(
			"HINCRBY",
			"battle:" + battleId + ":activePlayer",
			"actionsTaken",
			2,
		).Err
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}


}

//func (c *CharacterController) special(w http.ResponseWriter, r *http.Request) {
//
//}