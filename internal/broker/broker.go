package broker

import (
	"fmt"
	"log"
	"net/http"
)

type (
	// Broker pitää kirjaa avoimista yhteyksistä, ja välittää edelleen
	// Notifier-kanavaan tulevat viestit kaikille kiinnostuneille
	// kuunteleville asiakkaille.
	// Toteuttaa: https://developer.mozilla.org/en-US/docs/Web/API/EventSource
	Broker struct {
		// Sisäinen vastaanottopiste uusille viesteille
		Notifier chan []byte
		// Uusien asiakasyhteyksien kirjaus
		newClients chan chan []byte
		// Poistuvien asiakasyhteyksien kirjaus
		closingClients chan chan []byte
		// Luettelo kaikista rekisteröidyistä asiakkaista
		clients map[chan []byte]bool
	}
)

func (messenger *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Varmista että tyyppi 'rw':n takana tukee pakotettua tyhjennystä
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "GSI-dataa ei voida striimata, kirjastossa ei ole sille tukea!", http.StatusInternalServerError)
		return
	}

	// Aseta tapahtumien striimaamisen vaatiman HTTP-otsaketiedot
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	messenger.newClients <- messageChan

	defer func() {
		messenger.closingClients <- messageChan
	}()

	go func() {
		// Odottaa asiakasohjelman yhteyden sulkua ja välittää viestin edelleen kirjanpitoon
		<-req.Context().Done()
		messenger.closingClients <- messageChan
	}()

	// Asiakaskohtainen silmukka ja kanava odottaa aina uutta viestiä välittäjältä
	for {
		// Kirjoita välittäjältä saatu viesti HTTP vastausvirtaan
		// Server Sent Events -yhteensopiva. data-avaimen lisäksi
		// voisi olla myös event-avain (esim. "added", "deleted", tms.)
		fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)

		// Laita jonoon kirjoitettu viesti välittömästi eteenpäin
		flusher.Flush()
	}
}

// Kuuntele eri Go kanavia ja toimi niiden mukaisesti uusien viestien saapuessa
func (messenger *Broker) listen() {
	for {
		select {
		case s := <-messenger.newClients:
			// Uusi asiakas avasi yhteyden. Lisää sen kanava vastaanottajien listalle.
			messenger.clients[s] = true
			log.Printf("Asiakas lisätty. Rekiströityjä asiakkaita: %d", len(messenger.clients))

		case s := <-messenger.closingClients:
			// Asiakas sulki yhteytensä, lopeta viestien lähettäminen sen kanavaan
			delete(messenger.clients, s)
			log.Printf("Asiakas poistettu. Rekisteröityjä asiakkaita: %d", len(messenger.clients))

		case event := <-messenger.Notifier:
			// Uusi GSI-viesti vastaanotettu, välitä se kaikille asiakkaille eteenpäin!
			for clientMessageChan, _ := range messenger.clients {
				clientMessageChan <- event
			}
		}
	}
}

// Broker factory
func NewServer() (messenger *Broker) {
	// Alusta uusi viestinvälittäjä
	messenger = &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	// Käynnistä kuuntele-ja-välitä -silmukka
	go messenger.listen()

	return
}
