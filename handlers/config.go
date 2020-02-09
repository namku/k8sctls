// config.go
package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Cluster struct which contains a name
// a profile a region and a serialnumber
type Cluster struct {
	Name         string `json:"name"`
	Profile      string `json:"profile"`
	Region       string `json:"region"`
	SerialNumber string `json:"serial-number"`
}

// Clusters struct which contains
// an array of clusters
type Clusters struct {
	Clusters []Cluster `json:"clusters"`
}

// func NewConfig() {
// 	readJson()
// }

func NewConfig() *Clusters {
	// Open jsonFile
	jsonFile, err := os.Open("config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("json config %v", err)
	}

	fmt.Println("Successfully Opened config.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Clusters array
	var clusters Clusters

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'clusters' which we defined above
	json.Unmarshal(byteValue, &clusters)

	return &clusters
	// for c := 0; c < len(clusters.Clusters); c++ {
	// 	fmt.Println(clusters.Clusters[c].Name)
	// }
}
