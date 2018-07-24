package main

import (
	"github.com/jmoiron/sqlx"
	"net/http"
	"log"
	"encoding/json"
	"github.com/husobee/vestigo"
	"io/ioutil"
)

type Enemy struct {
	Id 			int				`json:"id"`
	Name 		string			`json:"name"`
	Level		int				`json:"level"`
	Exp			int				`json:"exp"`
	MaxHP		int				`json:"maxHp"`
	MaxMP		int				`json:"maxMp"`
	Strength	int				`json:"strength"`
	Speed		int				`json:"speed"`
	Move		int				`json:"move"`
	Jump 		int				`json:"jump"`
}

type EnemyController struct {
	DB		*sqlx.DB
}

func (c *EnemyController) getEnemies(w http.ResponseWriter, r *http.Request) {
	var enemies []Enemy
	err := c.DB.Select(&enemies, "SELECT * FROM enemy ORDER BY id")
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(enemies)
}

func (c *EnemyController) getEnemy(w http.ResponseWriter, r *http.Request) {
	var enemy Enemy
	id := vestigo.Param(r, "id")

	err := c.DB.Get(&enemy, "SELECT * FROM enemy WHERE id=$1", id)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(enemy)
}

func (c *EnemyController) createEnemy(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var enemy Enemy
	err = json.Unmarshal(b, &enemy)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var lastInsertId int;
	enemyInsert := `
		INSERT INTO enemy (
			name,
			level,
			exp,
			hp,
			mp,
			move,
			jump,
			speed,
		  ) VALUES (
		  	$1,
		  	$2,
		  	$3,
		  	$4,
		  	$5,
		  	$6,
		  	$7,
		  	$8,
		  ) RETURNING id
	`

	err = c.DB.QueryRow(
		enemyInsert,
		enemy.Name,
		0,
		0,
		20,
		5,
		3,
		3,
		5,
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