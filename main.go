package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func main() {
	configDirectory := "./bots"
	files, err := ioutil.ReadDir(configDirectory)
	if err != nil {
		log.Fatal(err)
	}

	r, _ := regexp.Compile(`\.json$`)
	for _, file := range files {
		if !file.IsDir() && r.FindString(file.Name()) != "" && file.Name() != "sample.json" {
			fileFullPath := fmt.Sprintf("%s/%s", configDirectory, file.Name())
			createTicker(fileFullPath)
		}
	}

	for {
		select {}
	}
}

func createTicker(jsonFilePath string) {
	log.Infoln("Loading ticker from: ", jsonFilePath)

	jsonBytes, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Errorf("Error reading json file: ", err)
	}

	ticker := Ticker{}
	err = json.Unmarshal(jsonBytes, &ticker)
	if err != nil {
		log.Errorf("Error loading ticker: ", err)
	}

	go ticker.run()
}
