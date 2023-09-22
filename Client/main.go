package main

import (
	"MECM2M-Emulator-Cloud/pkg/m2mapi"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	var data any
	var url string
	args := os.Args

	switch args[1] {
	case "area":
		data = m2mapi.AreaMapping{
			NE: m2mapi.SquarePoint{Lat: 35.531, Lon: 139.531},
			SW: m2mapi.SquarePoint{Lat: 35.53, Lon: 139.53},
		}
		url = "http://localhost:8080/m2mapi/area/mapping"
	default:
		fmt.Println("There is no args")
		log.Fatal()
	}

	client_data, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling data: ", err)
		return
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(client_data))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Server Response:", string(body))
}
