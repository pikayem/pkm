package steamid

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	ResolvedViaFailed = iota
	ResolvedViaVanityURL
	ResolvedViaID
	ResolvedViaID3
	ResolvedViaID32
	ResolvedViaID64
)

var (
	idRegex   = regexp.MustCompile(`^STEAM_(0|1):(0|1):[0-9]{1}[0-9]{0,8}$`)
	id3Regex  = regexp.MustCompile(`(\[)?U:1:\d+(\])?`)
	id64Regex = regexp.MustCompile(`^\d{17}$`)
)

// ResolveID attempts to resolve a SteamID 64 from a search query.
// It checks for queries such as vanity url's or SteamID's.
// If an invalid API key is used, a SteamID 64 may still be resolved if
// the query is not a vanity url.
// If no SteamID 64 could be resolved from the query, then 0 is returned.
// The second argument is a uint8 represeting how the SteamID 64 may have
// been resolved.
func ResolveID(query, apiKey string) (ID64, uint8) {
	query = strings.Replace(query, " ", "", -1)
	query = strings.Trim(query, "/")

	if strings.Contains(query, "steamcommunity.com/profiles/") {
		id64, err := strconv.ParseInt(query[strings.Index(query, "steamcommunity.com/profiles/")+len("steamcommunity.com/profiles/"):], 10, 64)
		if err != nil {
			goto isID
		}

		if len(strconv.FormatInt(id64, 10)) != 17 {
			goto isID
		}

		return steamID64(id64), ResolvedViaID64
	}

isID:
	if idRegex.MatchString(query) {
		id64 := steamID(query).To64()

		if len(strconv.FormatUint(id64.Uint64(), 10)) != 17 {
			goto isID3
		}

		return id64, ResolvedViaID
	}

isID3:
	if id3Regex.MatchString(strings.ToUpper(query)) {
		return steamID3(query).To64(), ResolvedViaID3
	}

	{
		if strings.Contains(query, "steamcommunity.com/id/") {
			query = query[strings.Index(query, "steamcommunity.com/id/")+len("steamcommunity.com/id/"):]
		}

		urlData := url.Values{
			"key":       {apiKey},
			"vanityurl": {query},
		}
		resp, err := http.Get("http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?" + urlData.Encode())
		if err != nil {
			goto isID64
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			goto isID64
		}

		var vanityURLJSON struct {
			Response struct {
				Steamid string
				Success int
			}
		}

		err = json.Unmarshal(body, &vanityURLJSON)
		if err != nil {
			goto isID64
		}

		if vanityURLJSON.Response.Success != 1 {
			goto isID64
		}

		if len(vanityURLJSON.Response.Steamid) != 17 {
			goto isID64
		}

		id64, err := strconv.ParseInt(vanityURLJSON.Response.Steamid, 10, 64)
		if err != nil {
			goto isID64
		}

		return steamID64(id64), ResolvedViaVanityURL
	}

isID64:
	if id64Regex.MatchString(query) {
		id64, err := strconv.ParseInt(query, 10, 64)
		if err != nil {
			goto isID32
		}

		return steamID64(id64), ResolvedViaID64
	}

isID32:
	id32, err := strconv.ParseInt(query, 10, 64)
	if err != nil {
		return steamID64(0), ResolvedViaFailed
	}

	if id32 >= 2 && id32 <= 4294967295 {
		return steamID32(id32).To64(), ResolvedViaID32
	}

	return steamID64(0), ResolvedViaFailed
}
