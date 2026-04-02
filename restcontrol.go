package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ResponseCellUpdateRequest struct {
	Contents GridCell `json:"contents"`
}

type ResponseCellUpdated struct {
	Message  string   `json:"message"`
	Status   int      `json:"status"`
	GridCell GridCell `json:"grid-cell"`
}

type ErrorResponse struct {
	Message  string   `json:"message"`
	Status   int      `json:"status"`
}

type MapListResponse struct {
	Message  string         `json:"message"`
	Status   int            `json:"status"`
	Maps     []MapListEntry `json:"maps"`
}

type MapMetadataResponse struct {
	Message  string         `json:"message"`
	Status   int            `json:"status"`
	Data     HexMapMetadata `json:"data"`
}

type MapResponse struct {
	Message  string   `json:"message"`
	Status   int      `json:"status"`
	Map      HexMap   `json:"map"`
}

type MapListEntry struct {
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Version  uint64 `json:"version"`
}

type HexMapMetadata struct {
	Filename string              `json:"filename"`
	Title   string               `json:"title"`
	Version uint64               `json:"version"`
	MinX int                     `json:"min_x"`
	MaxX int                     `json:"max_x"`
	MinY int                     `json:"min_y"`
	MaxY int                     `json:"max_y"`
}

func StartRestController(hexMaps map[string]HexMap, transformTasks chan MapTransformTask, port string) {
	// Define a route and handler for serving static files at the root
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", http.StripPrefix("/", fs))

	http.HandleFunc("/maps", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "GET" {
			res := ErrorResponse {
				Message: fmt.Sprintf("Bad request. Method not supported."),
				Status: 400,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		mapList := make([]MapListEntry, 0, len(hexMaps))
		for k, v := range hexMaps {
			entry := MapListEntry{
				Filename: k,
				Title:    v.Title,
				Version:  v.Version,
			}

			mapList = append(mapList, entry)
		}

		res := MapListResponse {
			Message:  "Success",
			Status:   200,
			Maps:     mapList,
		}
		json.NewEncoder(w).Encode(res)
	})

	http.HandleFunc("/maps/{name}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		if r.Method != "GET" {
			res := ErrorResponse {
				Message: fmt.Sprintf("Bad request. Method not supported."),
				Status: 400,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		mapName := r.PathValue("name")

		hexMap := hexMaps[mapName]

		hexMap, exists := hexMaps[mapName]
		if !exists {
			res := ErrorResponse {
				Message: fmt.Sprintf("Map '%s' does not exist.", mapName),
				Status: 404,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		res := MapResponse {
			Message:  "Success",
			Status:   200,
			Map:      hexMap,
		}
		json.NewEncoder(w).Encode(res)
	})

	http.HandleFunc("/maps/{name}/metadata", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		if r.Method != "GET" {
			res := ErrorResponse {
				Message: fmt.Sprintf("Bad request. Method not supported."),
				Status: 400,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		mapName := r.PathValue("name")

		hexMap := hexMaps[mapName]

		hexMap, exists := hexMaps[mapName]
		if !exists {
			res := ErrorResponse {
				Message: fmt.Sprintf("Map '%s' does not exist.", mapName),
				Status: 404,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		mapMeta := HexMapMetadata{
				Filename: mapName,
				Title:    hexMap.Title,
				Version:  hexMap.Version,
				MinX:     hexMap.MinX,
				MaxX:     hexMap.MaxX,
				MinY:     hexMap.MinY,
				MaxY:     hexMap.MaxY,
		}
		res := MapMetadataResponse {
			Message:  "Success",
			Status:   200,
			Data:     mapMeta,
		}
		json.NewEncoder(w).Encode(res)
	})

	http.HandleFunc("/maps/{name}/data/{x}/{y}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			res := ErrorResponse {
				Message: fmt.Sprintf("Bad request. Method not supported."),
				Status: 400,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		mapName := r.PathValue("name")
		x := r.PathValue("x")
		y := r.PathValue("y")

		var newCell ResponseCellUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&newCell)
		if err != nil {
			res := ErrorResponse {
				Message: fmt.Sprintf("Bad request. Bad JSON body submitted.", mapName),
				Status: 400,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		log.Printf("  - Updating grid cell (%s, %s)) {contents: '%s'}\n", x, y, newCell.Contents)

		res := ResponseCellUpdated{
			Message:  "Accepted",
			Status:   202,
			GridCell: newCell.Contents,
		}

		newHexMap := hexMaps[mapName]
		newHexMap.HexGrid[x+"/"+y] = newCell.Contents
		newHexMap.Version += 1
		hexMaps[mapName] = newHexMap

		saveMap(mapName, newHexMap)

		json.NewEncoder(w).Encode(res)
	})

	// Start the server on port 8080
	log.Println("Server starting on http://localhost:"+port+"...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
