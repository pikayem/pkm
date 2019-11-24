/*
	Package steamid is used for general SteamID conversion to and from any existing SteamID.
	It also currently supports resolving a SteamID 64 from a search query.


	Converting a SteamID

	To convert a SteamID 64 bit (ID64) to a regular SteamID (ID) for example, you would
	use the following code:

		// Define type the SteamID as type ID.
		id := steamid.NewID("STEAM_0:0:86173181")

		// Convert id64 to a regular SteamID.
		id64 := id64.ToID64()


	Finding an ID64 with a Search Query

	ResolveID is a function available in this package that attempts to resolve a SteamID 64 from a
	search query. It resolves queries by first identifying what the query is (eg. a vanity url) and
	completes multiple tasks in an attempt successfully resolve the query.

	You can use this function with a Steam web API key which is used to resolve the query
	as a vanity URL, or, if you do not specify a valid API key (an empty string), then in
	most cases you will still be able to find the ID64 as long as the search query is a
	SteamID in another format.

	Here is an example of the ResolveID function in use:

		steamid.ResolveID("http://steamcommunity.com/id/gabelogannewell", "xxx")

	Which would return 76561197960287930 as an ID64.
*/
package steamid

// ID is a regular SteamID. STEAM_0:0:86173181
type ID interface {
	String() string
	To64() ID64
	To32() ID32
	To3() ID3
}

// ID64 is a SteamID 64 bit. 76561198132612090
type ID64 interface {
	Uint64() uint64
	ToID() ID
	To32() ID32
	To3() ID3
}

// ID32 is a SteamID 32 bit. 172346362
type ID32 interface {
	Uint32() uint32
	ToID() ID
	To64() ID64
	To3() ID3
}

// ID3 is a SteamID v3. [U:1:172346362]
type ID3 interface {
	String() string
	ToID() ID
	To64() ID64
	To32() ID32
}

type (
	steamID   string
	steamID64 uint64
	steamID32 uint32
	steamID3  string
)

// NewID returns an ID from a string.
func NewID(id string) ID {
	return steamID(id)
}

// NewID64 returns an ID64 from a uint64.
func NewID64(id64 uint64) ID64 {
	return steamID64(id64)
}

// NewID32 returns an ID32 from a uint32.
func NewID32(id32 uint32) ID32 {
	return steamID32(id32)
}

// NewID3 returns an ID3 from a string.
func NewID3(id3 string) ID3 {
	return steamID3(id3)
}
