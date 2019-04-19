package internal

import (
	"bytes"
	"flag"
	"github.com/gorilla/mux"
	"github.com/jmoiron/jsonq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func Run() {
	setup()

	listenAddress := listenAddress()
	log.Print("PKM palvelin käynnistyy osoitteessa: " + listenAddress)

	router := mux.NewRouter()

	router.HandleFunc("/", ReceiveGameStatus)
	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

// ReceiveGameStatus käsittelee CS:GO observerin lähettämän pelidatapaketin
func ReceiveGameStatus(w http.ResponseWriter, r *http.Request) {
	var data *jsonq.JsonQuery
	data = DecodeJsonToJsonQ(bytes.NewReader(getRawPost(r)))

	// Varmista että JSON:issa tuli mukana pelaajatieto ja yritä vaihtaa kuvaa ainoastaan jos se löytyy
	player, err := data.Object("player")
	if err != nil {
		log.Println("GSI JSON player elementin lukeminen epäonnistui: ", err)
	}
	if player != nil {
		SwitchPlayer(player["steamid"].(string))
		log.Print("\"" + player["steamid"].(string) + "\": {\"player_name\": \"" + player["name"].(string) + "\", \"place\": 0},")
	}

	w.WriteHeader(http.StatusOK)
}

func getRawPost(r *http.Request) (body []byte) {
	var err error
	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func setup() {
	pConfFilename := flag.String("conf", "pkm.json", "JSON konfiguraatiotiedosto yleisille asetuksille")

	obsConfig := Config{}
	obsConfig.TeamAFile = flag.String("A", "team1.json", "JSON konfiguraatiotiedosto A-tiimille")
	obsConfig.TeamBFile = flag.String("B", "team2.json", "JSON konfiguraatiotiedosto B-tiimille")
	obsConfig.TestOnly = flag.Bool("test", false, "testaa palvelinsovellusta paikallisesti lähettämättä ohjauskomentoja")
	flag.Parse()

	ConfigurePKM(*pConfFilename)
	ConfigureOBS(obsConfig)
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
