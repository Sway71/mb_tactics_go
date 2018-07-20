package main

import (
	"github.com/jmoiron/sqlx"
	"net/http"
	"log"
	"encoding/json"
	"github.com/husobee/vestigo"
	"io/ioutil"
)

type MapTile struct {
	Height			int					`json:"height"`
	Terrain			string				`json:"terrain"`
}

type Map struct {
	Id				int					`json:"id"`
	Name			string				`json:"name"`
	MapData			json.RawMessage		`json:"mapData" db:"map_data"`
}

type MapController struct {
	DB		*sqlx.DB
}

func (m *MapController) getMaps(w http.ResponseWriter, r *http.Request) {
	var maps []Map
	err := m.DB.Select(&maps, "SELECT * FROM battlefield ORDER BY id")
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(maps)
}

func (m *MapController) getMap(w http.ResponseWriter, r *http.Request) {
	var singleMap Map
	id := vestigo.Param(r, "id")

	err := m.DB.Get(&singleMap, "SELECT * FROM battlefield WHERE id=$1", id)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(singleMap)
}

func (m *MapController) createMap(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var singleMap Map
	err = json.Unmarshal(b, &singleMap)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var mapDataArray [][]MapTile
	err = json.Unmarshal(singleMap.MapData, &mapDataArray)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// TODO: do some kind of validation on this mapDataArray so we know the JSON that goes into the db is clean

	var lastInsertId int;
	battlefieldInsert := `
		INSERT INTO battlefield (
			name,
			map_data
		  ) VALUES (
		  	$1,
		  	$2
		  ) RETURNING id
	`
	err = m.DB.QueryRow(
		battlefieldInsert,
		singleMap.Name,
		singleMap.MapData,
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