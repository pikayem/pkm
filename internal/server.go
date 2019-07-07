package internal

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"github.com/jmoiron/jsonq"
	"github.com/pikayem/pkm/internal/broker"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	teams       map[string]map[string]Player
	lastGSIJSON []byte
	messenger   *broker.Broker
)

func Run() {
	setup()

	pkmListenAddress := pkmListenAddress()
	log.Print("PKM palvelin käynnistyy osoitteessa: " + pkmListenAddress)

	router := mux.NewRouter()

	router.HandleFunc("/", ReceiveGameStatus)
	router.HandleFunc("/state", ReportGameState)
	router.HandleFunc("/players", ReportConfPlayers).Methods("GET", "OPTIONS")
	router.HandleFunc("/lastgsijson", ReportLastGSIJSON)
	//http.Handle("/", router)

	go func() {
		log.Fatal(http.ListenAndServe(pkmListenAddress, nil))
	}()

	gsiPusherListenAddress := gsiPusherListenAddress()
	log.Print("GSI Pusher palvelin käynnistyy osoitteessa: " + gsiPusherListenAddress)

	messenger = broker.NewServer()
	log.Fatal("GSI välityspalvelun virhe: ", http.ListenAndServe(gsiPusherListenAddress, messenger))
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

	s, err := json.Marshal(teams)
	if err != nil {
		log.Println("Joukkuestatuksen JSON-käännös epäonnistui: ", err)
	}

	messenger.Notifier <- []byte(s)
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

func pkmListenAddress() string {
	var address, port string
	var err error

	address, err = CQ.String("pkm", "address")
	if err != nil {
		log.Fatal("Puuttuva tai virheellinen PKM osoitekonfiguraatio: ", err)
	}

	port, err = CQ.String("pkm", "port")
	if err != nil {
		log.Fatal("Puuttuva tai virheellinen PKM porttikonfiguraatio: ", err)
	}

	return address + ":" + port
}

func gsiPusherListenAddress() string {
	var address, port string
	var err error

	address, err = CQ.String("gsi_pusher", "address")
	if err != nil {
		log.Fatal("Puuttuva tai virheellinen GSI Pusher osoitekonfiguraatio: ", err)
		os.Exit(1)
	}

	port, err = CQ.String("gsi_pusher", "port")
	if err != nil {
		log.Fatal("Puuttuva tai virheellinen GSI Pusher porttikonfiguraatio: ", err)
		os.Exit(1)
	}

	return address + ":" + port
}
