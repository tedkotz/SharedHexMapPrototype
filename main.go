package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type MapTransformTask struct {
	MapId     string
	Cell      GridCell
	Timestamp time.Time
}

var transformTaskChannel chan MapTransformTask


func mapTransformWorker() {
	for task := range transformTaskChannel {
		fmt.Printf("Worker processing task: %s\n", task.Timestamp)
		// Simulate some work
		// time.Sleep(time.Second)
		log.Printf("Worker finished job: %s\n", task.Timestamp)
	}
}

func saveMap(jsonFileName string, hexmap HexMap) (error) {
	jsonFilePath := filepath.Join(".","maps", jsonFileName)
	log.Printf("- Saving '%s'\n", jsonFilePath)

	byteValue, err := json.MarshalIndent( hexmap, "", "  " )
	if err != nil {
		log.Println("Error Marshaling JSON:", err)
		return err
	}

	backupFilePath := jsonFilePath + "." + strconv.FormatUint(hexmap.Version - 1, 10) + ".bak"

	err = os.Rename(jsonFilePath, backupFilePath)
	if err != nil {
		log.Println("Error making backup JSON:", err)
		return err
	}


	err = os.WriteFile(jsonFilePath, byteValue, 0644)
	if err != nil {
		log.Println("Error Writing JSON to file:", err)
		return err
	}

	log.Println("Successfully saved.")

	return nil
}

func loadMap(jsonFilePath string) (HexMap, error) {
	log.Printf("- Loading '%s'\n", jsonFilePath)
	var hexmap HexMap

	jsonFile, err := os.Open(jsonFilePath)
	// if os.Open returns an error then handle it
	if err != nil {
		return hexmap, err
	}

	log.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return hexmap, err
	}

	err = json.Unmarshal(byteValue, &hexmap)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return hexmap, err
	}

	return hexmap, nil
}

func main() {
	// TODO: Add/check/remove a lockfile with ./maps/hexmap.lock.${pid} to
	// make sure only one instance of this program is running at a time.
	// defer (); signal.NotifyContext ; os.Getpid

	// Create in memory data store for all maps
	hexMaps := make(map[string]HexMap)

	files, err := filepath.Glob(filepath.Join(".", "maps","*.json"))
	if err != nil {
		log.Fatal(err)
	}

	for _, fullPath := range files {
		hexMap, err := loadMap(fullPath)
		if err != nil {
			log.Fatal(err)
		}

		hexMaps[filepath.Base(fullPath)] = hexMap
	}

	// Create go channel for map transform operations to be created by the web service
	// and consumed by the consumer routine
	transformTaskChannel := make(chan MapTransformTask)

	go mapTransformWorker()
	if len(os.Args) > 1 {
		StartRestController(hexMaps, transformTaskChannel, os.Args[1])
	} else {
		StartRestController(hexMaps, transformTaskChannel, "8080")
	}
}
