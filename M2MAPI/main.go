package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"mecm2m-Emulator-Cloud/pkg/m2mapi"

	"github.com/joho/godotenv"
)

var port string
var ip_address string
var rtt_between_machines [][]string

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

	file, err := os.Open(os.Getenv("HOME") + os.Getenv("PROJECT_NAME") + "/rtt_between_machines.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rtt_between_machines, _ = reader.ReadAll()
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

	log.Printf("Connect to http://%s:%s/ for M2M API", ip_address, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func resolveAreaMapping(w http.ResponseWriter, r *http.Request) {
	// 初めに，リクエストを送信した送信元IPアドレスを確認して，RTTを模擬的に表現
	executeRTT(r)

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
		// 少しでも範囲が被っていれば，対象領域とする
		for _, cover_area := range cover_areas.MecServers {
			if ((inputFormat.NE.Lat <= cover_area.MaxLat && inputFormat.NE.Lat > cover_area.MinLat) || (inputFormat.SW.Lat >= cover_area.MinLat && inputFormat.SW.Lat < cover_area.MaxLat)) && ((inputFormat.NE.Lon <= cover_area.MaxLon && inputFormat.NE.Lon > cover_area.MinLon) || (inputFormat.SW.Lon >= cover_area.MinLon && inputFormat.SW.Lon < cover_area.MaxLon)) {
				// 対象領域
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

func executeRTT(r *http.Request) {
	var rtt time.Duration
	src_ip_address := strings.Split(r.RemoteAddr, ":")[0]
	//fmt.Println(src_ip_address)
	if src_ip_address == "127.0.0.1" {
		return
	} else {
		for _, rttComb := range rtt_between_machines {
			if (rttComb[0] == src_ip_address && rttComb[1] == ip_address) || (rttComb[1] == src_ip_address && rttComb[0] == ip_address) {
				rtt_float, _ := strconv.ParseFloat(rttComb[2], 64)
				rtt_str := strconv.FormatFloat(rtt_float, 'f', 2, 64) + "ms"
				rtt, _ = time.ParseDuration(rtt_str)
			}
		}
		time.Sleep(rtt)
	}
}
