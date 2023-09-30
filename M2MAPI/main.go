package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"mecm2m-Emulator-Cloud/pkg/m2mapi"

	"github.com/joho/godotenv"
)

var port string
var ip_address string

type MECCoverAreas struct {
	MecServers []m2mapi.MECCoverArea `json:"mec-servers"`
}

func init() {
	// .envファイルの読み込み
	if err := godotenv.Load(os.Getenv("HOME") + "/.env"); err != nil {
		log.Fatal(err)
	}
	port = os.Getenv("M2M_API_PORT")
	ip_address = os.Getenv("IP_ADDRESS")
}

func main() {
	/*
		// Mainプロセスのコマンドラインからシミュレーション実行開始シグナルを受信するまで待機
		signals_from_main := make(chan os.Signal, 1)

		// 停止しているプロセスを再開するために送信されるシグナル，SIGCONT(=18)を受信するように設定
		signal.Notify(signals_from_main, syscall.SIGCONT)

		// シグナルを待機
		fmt.Println("Waiting for signal...")
		sig := <-signals_from_main

		// 受信したシグナルを表示
		fmt.Printf("Received signal: %v\n", sig)
	*/
	http.HandleFunc("/m2mapi/area/mapping", resolveAreaMapping)
	http.HandleFunc("/hello", hello)

	log.Printf("Connect to http://%s:%s/ for M2M API", ip_address, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func resolveAreaMapping(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolvePoint: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.AreaMapping{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolvePoint: Error missmatching packet format", http.StatusInternalServerError)
		}

		// AreaMapping/mec_cover_area.jsonから入力であるSW, NEに該当するサーバIP群を探す．
		mec_cover_area := os.Getenv("HOME") + os.Getenv("PROJECT_NAME") + "/Area_Mapping/mec_cover_area.json"
		file, err := os.Open(mec_cover_area)
		if err != nil {
			fmt.Println("Error open json file: ", err)
			return
		}
		defer file.Close()

		var cover_areas MECCoverAreas
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&cover_areas)
		if err != nil {
			fmt.Println("Error decode json type: ", err)
			return
		}

		var results []m2mapi.AreaMapping
		for _, cover_area := range cover_areas.MecServers {
			if (inputFormat.NE.Lat <= cover_area.MinLat || inputFormat.NE.Lon <= cover_area.MinLon) || (inputFormat.SW.Lat >= cover_area.MaxLat || inputFormat.SW.Lon >= cover_area.MaxLon) {
				// 対象領域でない
				//fmt.Printf("(%f < %f && %f < %f) || (%f > %f && %f > %f)", inputFormat.NE.Lat, cover_area.MinLat, inputFormat.NE.Lon, cover_area.MinLon, inputFormat.SW.Lat, cover_area.MaxLat, inputFormat.SW.Lon, cover_area.MaxLon)
			} else {
				// 対象領域である
				fmt.Println("target area: ", cover_area.ServerIP)
				result := m2mapi.AreaMapping{}
				result.MECCoverArea.MinLat = cover_area.MinLat
				result.MECCoverArea.MaxLat = cover_area.MaxLat
				result.MECCoverArea.MinLon = cover_area.MinLon
				result.MECCoverArea.MaxLon = cover_area.MaxLon
				result.MECCoverArea.ServerIP = cover_area.ServerIP
				results = append(results, result)
			}
		}

		results_str, err := json.Marshal(results)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return
		}
		fmt.Fprintf(w, "%v\n", string(results_str))
	} else {
		http.Error(w, "resolvePoint: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World\n")
}
