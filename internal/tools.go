package internal

import (
	"encoding/json"
	"fmt"
	"github.com/Acidic9/go-steam/steamid"
	"github.com/jmoiron/jsonq"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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
	configuration := map[string]interface{}{}
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatalf("Konfiguraation lukuvirhe: %s", err)
		return nil
	}
	return jsonq.NewQuery(configuration)

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
