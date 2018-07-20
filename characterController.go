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
	X 			int				`json:"x"`
	Y 			int				`json:"y"`
}

type CharacterController struct {
	DB		*sqlx.DB
}

func (c *CharacterController) getCharacters(w http.ResponseWriter, r *http.Request) {
	characters := []Character{}
	err := c.DB.Select(&characters, "SELECT * FROM character ORDER BY id")
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(characters)
}

func (c *CharacterController) getCharacter(w http.ResponseWriter, r *http.Request) {
	character := Character{}
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
	character := Character{}
	id := vestigo.Param(r, "id")

	err := c.DB.Get(&character, "SELECT move, x, y FROM character WHERE id=$1", id)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(GetMovableSpaces(character.Move, Location{character.X, character.Y}))
}

func (c *CharacterController) move(w http.ResponseWriter, r *http.Request) {
	character := Character{}
	id := vestigo.Param(r, "id")

	err := c.DB.Get(&character, "SELECT move, x, y FROM character WHERE id=$1", id)
	if err != nil {
		log.Fatalln(err)
	}

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

	validMove := ContainsPoint(GetMovableSpaces(character.Move, Location{character.X, character.Y}), location)
	fmt.Println(validMove)

	if (validMove) {
		characterMove := "UPDATE character SET x = $1, y = $2 WHERE id = $3 RETURNING x, y"
		result := c.DB.MustExec(characterMove, location.X, location.Y, id)
		rowsAffected, _ := result.RowsAffected()

		wasSuccessful := rowsAffected == 1
		json.NewEncoder(w).Encode(struct {
			Success bool `json:"success"`
		}{
			wasSuccessful,
		})
	} else {
		json.NewEncoder(w).Encode(struct {
			Success bool `json:"success"`
		}{
			false,
		})
	}
}
//
//func (c *CharacterController) attack(w http.ResponseWriter, r *http.Request) {
//
//}
//
//func (c *CharacterController) special(w http.ResponseWriter, r *http.Request) {
//
//}