# Pelaajakuvamasiina
Golla kirjoitettu liima Counter Strike: Global Offensiven kuunteluun ja OBS:n ohjaamiseen halutun USB-liitäntäisen pelaajakameran valitsemiseksi.

Kaksi videoserveriä lähettävät kuvaa NDI:llä vmixille tms kuvamikserille, alphakanavan kera siten että vain akviinen kamera näkyy, jommalta kummalta serveriltä. Nämä voidaan sitten aktivoida PIP:nä tai muuten halutusti.

# Tarkoituksenmukainen kytkentä
```
[Pelaaja #1] ───┐
[Pelaaja #2] ───┤
[Pelaaja #3] ───┼ USB ─→┃videoserveri A┠── NDI ─┐
[Pelaaja #4] ───┤               ↑               │
[Pelaaja #5] ───┘               │               │
                            websocket           │
                                │               ↓
┃Observer┠─────── JSON ───→┃PKM kone┃         [vMix]
                                │               ↑
                            websocket           │
[Pelaaja #6] ───┐               │               │
[Pelaaja #7] ───┤               ↓               │
[Pelaaja #8] ───┼ USB ─→┃videoserveri B┠── NDI ─┘
[Pelaaja #9] ───┤
[Pelaaja #10] ──┘
```

# Riippuvuudet ja asennus
1. OBS studio [https://obsproject.com/](https://obsproject.com/)
2. NDI-plugin [https://github.com/Palakis/obs-ndi](https://github.com/Palakis/obs-ndi)
3. NDI redist [http://new.tk/NDIRedistV3](http://new.tk/NDIRedistV3)
4. obs-websocket plugin OBS:lle [https://github.com/Palakis/obs-websocket](https://github.com/Palakis/obs-websocket), >=4.3.0

Kummallekin videoserverille:
  * Asenna ylläolevat.
  * Avaa/luo tyhjä OBS scene collection. Tarvitaan yksi scene, nimeltään "Scene1"..
  * Aktivoi scenelle NDI dedicated output, suositeltavaa nimetä ne esim. "serverA" ja "serverB". Tämä on se nimi jolla kuvalähteet näkyvät NDI:n ylitse.

Kaikki videoserveriin kytketyt kamerat tuodaan tähän sceneen nimettynä "A1"-"A5" ensimmäisellä ja "B1"-"B5" toisella serverillä. Kameroiden numerointi kannattaa aloittaa pelaajien selän takaa katsoen vasemmalta. Kaikki kamerat asetetaan `fit-to-screen` tilaan, ja normaalitilassa kaikkien `visibility` pois päältä. Myös webbikameroiden kuva-/videoasetukset kannattaa tarkistaa optimaalisen kuvanlaadun saamiseksi. Serverin asennus pelipöytiin niin että A on SL, B SR puolella.

OBS:n Websocket-plugin kuuntelee oletuksena portissa 4444. Tällä hetkellä tässä projektissa ei ole autentikaatiotukea, koska järjestelmä on tarkoitettu vain suljetussa verkossa ajettavaksi.

Observer-koneelle asennetaan kansioon `steamapps\common\Counter-Strike Global Offensive\csgo\cfg` GSI-asetustiedosto (ks. `configs/gamestate_integration_pkm.cfg`). Pelin pitää pyöriä samassa verkossa tai palomuurissa pitää olla aukko peliverkosta PKM-koneen websocket-porttiin (oletus 1999).

[Lisätiedot GSI:stä.](https://developer.valvesoftware.com/wiki/Counter-Strike:_Global_Offensive_Game_State_Integration)

Asetustiedostoihin laitetaan pelaajien steamID:t SteamID, SteamID3, SteamID32 tai SteamID64 muodossa. Tiedostoja on yksi per joukkue. Paikat myöskin pelaajien takaa vasemmalta laskien. Paikka `0` tarkoittaa sitä, että pelaajalla ei ole kameraa tai kamera on esimerkiksi väärin suunnattu, ja sen takia halutaan hetkellisesti poistaa käytöstä näin:

  * editoi tiedostoa ja muuta halutulle kameralle paikaksi `0`,
  * keskeytä ajossa oleva ohjelma `ctrl+c` ja
  * käynnistä ohjelma uudelleen.
  
Mikäli PKM-kone on kytketty internettiin reitittävään verkkoon, voit lisätä ```pkm.exe```:n kanssa samaan kansioon myös ```steam.apikey``` tiedoston, jonka ainoa sisältö on yksi Steam Web API -avain. Tällöin PKM kysyy Steamilta konfiguraatioista lukemiaan SteamID:itä vastaavat pelaajien näyttönimet, tai raportoi jos jollain SteamID:llä ei löytynyt pelaajan tietoja Steamista.
  
API-avaimen saa Steam-tunnuksilla [Steamin kehittäjäportaalista](https://steamcommunity.com/dev/apikey), joka myös näyttää jo mahdollisetn aikaisemmin luodun avaimen kirjauduttuasi.

# Serverin käynnistys

Kopioi ja muokkaa `pkm.json`, `team1.json` ja `team2.json` tiedostot `pkm.exe`:n kanssa samaan hakemistoon. Sen jälkeen suorita:

`./pkm -A team2.json -B team1.json`

Joukkuekonfiguraatiot kannattaa kirjoittaa hyvissä ajoin etukäteen, jolloin PKM:n uudelleenkonfigurointi pelistä toiseen sujuu helposti vain PKM:n uudelleenkäynnistämällä uusilla joukkuetiedostoparametreilla.

PKM:n oman konfiguraation voi myös määrittää asuvan eri paikassa ```-conf``` vivulla.

# Rajapinnat

Järjestelmä osaa antaa tilatietoa ulospäin muille järjestelmille

* ```/state``` sisältää JSON-olion tällä hetkellä serverillä nähdyistä id:istä
* ```/players``` näyttää tällä hetkellä konfiguraatiosta ladatut pelaajat 
* ```/lastgsijson``` antaa istumapaikkatiedolla rikastetun GSI-datan
