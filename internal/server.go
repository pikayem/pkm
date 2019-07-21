package internal

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"github.com/jmoiron/jsonq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	teams map[string]map[string]Player
	lastGSIJSON []byte
)

func Run() {
	setup()

	listenAddress := listenAddress()
	log.Print("PKM palvelin käynnistyy osoitteessa: " + listenAddress)

	router := mux.NewRouter()

	router.HandleFunc("/", ReceiveGameStatus)
	router.HandleFunc("/state", ReportGameState)
	router.HandleFunc("/players", ReportConfPlayers).Methods("GET", "OPTIONS")
	router.HandleFunc("/lastgsijson", ReportLastGSIJSON)
	//http.Handle("/", router)

	log.Fatal(http.ListenAndServe(listenAddress, router))
}

// ReceiveGameStatus käsittelee CS:GO observerin lähettämän pelidatapaketin
func ReceiveGameStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var data *jsonq.JsonQuery
	raw := getRawPost(r)
	lastGSIJSON = raw
	data = DecodeJsonToJsonQ(bytes.NewReader(raw))

	_ = updateGameState(data)
	_ = updateObserverState(data)

	w.WriteHeader(http.StatusOK)
}

func ReportGameState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	s, err := json.MarshalIndent(teams, "", "    ")
	if err != nil {
		log.Println("Joukkuestatuksen JSON-käännös epäonnistui: ", err)
	}
	w.Write(s)
}

func ReportConfPlayers(w http.ResponseWriter, r *http.Request) {
	//w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	s, err := json.MarshalIndent(Players, "", "    ")
	if err != nil {
		log.Println("Pelaajaconfin JSON-käännös epäonnistui: ", err)
	}
	w.Write(s)
}

func ReportLastGSIJSON(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusOK)
	w.Write(lastGSIJSON)
}

func getRawPost(r *http.Request) (body []byte) {
	var err error
	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func updateObserverState(data *jsonq.JsonQuery) error {
	// Varmista että JSON:issa tuli mukana pelaajatieto ja yritä vaihtaa kuvaa ainoastaan jos se löytyy
	player, err := data.Object("player")
	if err != nil {
		log.Println("GSI JSON player elementin lukeminen epäonnistui: ", err)
	}
	if player != nil {
		SwitchPlayer(player["steamid"].(string))
		log.Print("Observattavana: \"" + player["steamid"].(string) + "\": {\"player_name\": \"" + player["name"].(string) + "\", \"place\": 0},")
	}

	return err
}

func updateGameState(data *jsonq.JsonQuery) error {
	allPlayers, err := data.Object("allplayers")
	if err != nil {
		log.Println(err)
	} else {
		for k, v := range allPlayers {
			p := Player{v.(map[string]interface{})["name"].(string), "", 0}
			switch v.(map[string]interface{})["team"].(string) {
			case "T":
				teams["T"][k] = p
			case "CT":
				teams["CT"][k] = p
			}
		}
	}

	return err
}

func setup() {
	pConfFilename := flag.String("conf", "pkm.json", "JSON konfiguraatiotiedosto yleisille asetuksille")

	obsConfig := Config{}
	obsConfig.TeamAFile = flag.String("A", "", "JSON konfiguraatiotiedosto A-tiimille")
	obsConfig.TeamBFile = flag.String("B", "", "JSON konfiguraatiotiedosto B-tiimille")
	obsConfig.TestOnly = flag.Bool("test", false, "testaa palvelinsovellusta paikallisesti lähettämättä ohjauskomentoja")
	flag.Parse()

	configureGameState()
	ConfigurePKM(*pConfFilename)
	ConfigureOBS(obsConfig)
}

func configureGameState() {
	teams = make(map[string]map[string]Player)
	teams["T"] = make(map[string]Player)
	teams["CT"] = make(map[string]Player)
}

func listenAddress() string {
	var address, port string
	var err error

	address, err = CQ.String("pkm", "address")
	if err != nil {
		log.Fatal("Puuttuva tai virheellinen PKM osoitekonfiguraatio: ", err)
		os.Exit(1)
	}

	port, err = CQ.String("pkm", "port")
	if err != nil {
		log.Fatal("Puuttuva tai virheellinen PKM porttikonfiguraatio: ", err)
		os.Exit(1)
	}

	return address + ":" + port
}
