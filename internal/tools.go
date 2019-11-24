package internal

import (
	"encoding/json"
	"fmt"
	"github.com/Acidic9/go-steam/steamid"
	"github.com/jmoiron/jsonq"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	apikeyFilename                = "steam.apikey"
	steamWebAPIGetPlayerSummaries = "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/"
)

var (
	CQ *jsonq.JsonQuery
)

func ConfigurePKM(filename string) {
	CQ = LoadJsonFile(filename)
}

func LoadJsonFile(filename string) *jsonq.JsonQuery {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("JSON-tiedoston %s lataaminen epäonnistui: %s", filename, err)
		return nil
	}
	return DecodeJsonToJsonQ(file)
}

func DecodeJsonToJsonQ(reader io.Reader) *jsonq.JsonQuery {
	var err error
	decoder := json.NewDecoder(reader)
	jsonStructure := map[string]interface{}{}
	err = decoder.Decode(&jsonStructure)
	if err != nil {
		log.Fatalf("JSON-rakenteen lukuvirhe: %s", err)
		return nil
	}
	return jsonq.NewQuery(jsonStructure)
}

func UnifySteamId(confSteamId string) string {
	// Yhdenmukaista SteamID, SteamID3 tai SteamID32 SteamID64 muotoon
	var steamId64 steamid.ID64
	var err error

	switch strings.ToUpper(string([]rune(confSteamId)[0])) {

	// SteamID
	case "S":
		steamId64 = steamid.NewID(confSteamId).To64()
		break

	// SteamID3, salli vain yksittäisen käyttäjän ID-tyyppi ("U")
	case "U":
		steamId64 = steamid.NewID3(confSteamId).To64()
		break

	// SteamID32 tai SteamID64
	default:
		var intermediateConfSteamId int

		intermediateConfSteamId, err = strconv.Atoi(confSteamId)
		if err != nil {
			log.Fatalf("SteamID '%s' näytti SteamID32:lta tai SteamID64:lta, "+
				"mutta kokonaisluvuksi muuttaminen epäonnistui: %s", confSteamId, err)
		}

		if len([]rune(confSteamId)) < 11 {
			steamId64 = steamid.NewID32(uint32(intermediateConfSteamId)).To64()
		} else {
			steamId64 = steamid.NewID64(uint64(intermediateConfSteamId))
		}
	}

	return strconv.Itoa(int(steamId64.Uint64()))
}

func VerifySteamId(steamId string) bool {
	var apikey []byte
	var err error

	apikey, err = ioutil.ReadFile(apikeyFilename)
	if err != nil {
		log.Printf("SteamID:tä %d ei tarkistettu, Steam API-avaintiedostoa '%s' ei voitu avata: %s", steamId, apikeyFilename, err)
		return false
	}
	var resp *http.Response
	queryUrl := fmt.Sprintf("%s?key=%s&steamids=%s", steamWebAPIGetPlayerSummaries, string(apikey), steamId)

	resp, err = http.Get(queryUrl)
	if err != nil {
		log.Fatalf("HTTPS GET Steam API:in epäonnistui: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		steamUser := DecodeJsonToJsonQ(resp.Body)
		players, err := steamUser.Array("response", "players")
		if err != nil {
			log.Fatalf("Players-listaa ei voitu parsia Steam API-kutsun vastauksesta: %s", err)
		}

		switch len(players) {
		case 0:
			log.Printf("SteamID:tä %s ei löytynyt Steamista", steamId)
			return false
		case 1:
			name, err := steamUser.String("response", "players", "0", "personaname")
			if err != nil {
				log.Printf("Steam-käyttäjänimen parsinta JSON-vastauksesta epäonnistui: %s", err)
				return  false
			}
			log.Printf("SteamID %s, käyttäjätunnus %s", steamId, name)
			return true
		default:
			log.Println("Liian monta pelaajatietuetta palautettu Steamista")
			return false
		}
	} else {
		log.Printf("API-kutsun suoritus palautti virheen, SteamID:tä ei voitu tarkistaa: %s", resp.StatusCode)
	}
	return false
}
