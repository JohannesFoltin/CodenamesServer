package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var isBlueGC bool
var isRedGC bool
var game Game

//Typen: GCB,GCR,AgentB,AgentR
type User struct {
	ws       *websocket.Conn `json:"-"`
	Name     string
	Typ      string
	Selected string
}

// Who Owns It: Red,Blue,Gray,Black

type Card struct {
	Word   string
	Owner  string
	Coverd bool
}

// Red == false, blue == true
type Game struct {
	Code        string
	Picks       int
	CurrentTeam bool
	Cards       []Card
	Users       []User
}

type Message struct{
    Goal string
    ParamOne string
    ParamTwo int
}

func gameInit() {
	startingPick := false
	if rand.Intn(2) == 0 {
		startingPick = true
	}
	game = Game{"", 0, startingPick, nil, make([]User, 0)}
	cards := make([]Card, 0)
	rand.Shuffle(len(words), func(i int, j int) {
		words[i], words[j] = words[j], words[i]
	})
	for i := 0; i < 25; i++ {
		switch {
		case i < 9:
			if startingPick {
				cards = append(cards, Card{words[i], "Blue", true})
			} else {
				cards = append(cards, Card{words[i], "Red", true})
			}

		case 8 < i && i < 17:
			if !startingPick {
				cards = append(cards, Card{words[i], "Blue", true})
			} else {
				cards = append(cards, Card{words[i], "Red", true})
			}

		case 16 < i && i < 24:
			cards = append(cards, Card{words[i], "Grey", true})

		case i == 24:
			cards = append(cards, Card{words[i], "Black", true})
		}
	}
	game.Cards = cards
}

func cardSelection(u User, c string) {
	for i := 0; i < len(game.Users); i++ {
		if u == game.Users[i] {
			game.Users[i].Selected = c
		}
	}
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func searchSlice(slice []string, s string) int {
	tmp := -1
	for i := 0; i < len(s); i++ {
		if slice[i] == s {
			tmp = i
		}
	}
	return tmp
}
func broadcastGameState() {
	tmp, err := json.Marshal(game)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(game.Users); i++ {
		game.Users[i].ws.WriteMessage(1, []byte(tmp))
	}
}
func wsEndpoint(w http.ResponseWriter, r *http.Request) {

	//checkt die Herkunft von Clients die sich verbinden wollen. Return true, da wir nicht wählerisch sind. :)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")

	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}

	newUser := User{ws, "Unknown", "AgentRed", ""}
	game.Users = append(game.Users, newUser)

	reader(&game.Users[len(game.Users)-1])
}

func reader(u *User) {
	init := false
	for {
		// read in a message
		_, p, err := u.ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// print out that message for clarity
		tmpS := string(p)
		fmt.Println(tmpS)

		if !init {
			//Verification JSON: {"name": "Johannes","typ": "AgentB"}
			if tmpS[0] != '{' {
				u.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1003, "Unsupported Data. Please send Verification JSON"))
			} else {
				if (isBlueGC && u.Typ == "GCB") || (isRedGC && u.Typ == "GCR") {
					u.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1003, "This Type of Player is already taken. Please choose another one!"))
				}
				err := json.Unmarshal(p, u)
				fmt.Println(err)
				init = true
				fmt.Println(u.Name)
			}
		} else {
            tmpMssg := Message{"","",0}
			err := json.Unmarshal(p,&tmpMssg)
            if err != nil {
                fmt.Println(err)
            }
            switch{
            case tmpMssg.Goal == "Get":
                broadcastGameState()
            case tmpMssg.Goal == "Select":
                cardSelection(*u,tmpMssg.ParamOne)
            case tmpMssg.Goal == "Announce":

                game.Code = tmpMssg.ParamOne
                game.Picks = tmpMssg.ParamTwo
            }

		}

	}
}

// & Referenz mitgeben (Pointer erstellen)
// * wieder nur ein Element drauß machen

func main() {
	fmt.Println("Hello World")
	gameInit()
	//For "better" Random :)
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/ws", wsEndpoint)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var words = []string{"Melone", "Laster", "Peitsche", "Berliner", "Wal", "Leiter", "Hubschrauber", "Dame", "Horn", "Platte", "Forscher", "Hering", "Nacht", "Stift", "Apfel", "Mangel", "Barren", "Kiwi", "Soldat", "Bach", "Mikroskop", "Erde", "Kohle", "Mund", "Loge", "Kreuz", "Chor", "Wald", "Meer", "Flügel", "Staat", "Futter", "Pistole", "Abgabe", "Flöte", "Kunde", "England", "Pension", "Umzug", "Lehrer", "Bett", "Viertel", "Wand", "Rute", "Tempo", "Elfenbein", "Seite", "Aufzug", "Prinzessin", "Bar", "Welle", "Film", "Kamm", "Stern", "Blatt", "Haupt", "Europa", "Schneemann", "Watt", "Loch", "Lösung", "Shakespeare", "Schloss", "Birne", "Fisch", "Bauer", "Mutter", "Netz", "Gang", "Gift", "Strom", "Absatz", "Satz", "Löwe", "Bulle", "Turm", "Blinker", "Ketchup", "Mine", "Engel", "Wasser", "Niete", "Kanal", "Maler", "Indien", "Mittel", "Bär", "Dietrich", "Rasen", "Hollywood", "Pilot", "Schelle", "Mini", "Fleck", "Note", "Maschine", "Erika", "Burg", "Mandel", "Pol", "Drucker", "Karotte", "Tisch", "Star", "Schnabeltier", "Brand", "Gericht", "Schnur", "Rost", "New York", "Schiff", "Stempel", "Scheibe", "Hund", "Nuss", "Fackel", "Gut", "Knopf", "Eis", "Börse", "Teleskop", "Römer", "Busch", "Ring", "Akt", "Grad", "Honig", "Frankreich", "Pass", "Verband", "Boxer", "Linie", "Wanze", "Zahn", "Feige", "Auto", "Australien", "Mühle", "Auflauf", "Dichtung", "Takt", "Gold", "Kreis", "Harz", "Raute", "Zoll", "Linse", "Limousine", "Dinosaurier", "Optik", "Demo", "Stuhl", "Kirche", "Bremse", "Zelle", "Brötchen", "Brücke", "Pfeife", "Daumen", "Leuchte", "Rücken", "Ball", "Mast", "König", "Käfer", "Glas", "Flur", "Hand", "Spiel", "Washington", "Steuer", "Ninja", "Krebs", "Rom", "Messe", "Zentaur", "Decke", "Herz", "Essen", "Hupe", "Funken", "Maus", "Orange", "Atlantis", "Ritter", "Chemie", "Ton", "Superheld", "Kraft", "Luxemburg", "Pyramide", "Vorsatz", "Gabel", "Fliege", "Pinguin", "Siegel", "Koks", "Roboter", "Krieg", "Alpen", "Fall", "Strand", "Stock", "Koch", "Hut", "Fläche", "Jet", "Inka", "Tokio", "Toast", "Messer", "Bock", "art", "Spinne", "Adler", "Antarktis", "Känguruh", "Pflaster", "Bein", "Krone", "Loch Ness", "Konzert", "Tafel", "Spion", "Strauß", "Wind", "Weide", "Moskau", "Griechenland", "Rock", "Finger", "Flasche", "Bahn", "Krankheit", "Fest", "Tor", "Kiefer", "Schirm", "Fallschirm", "Stamm", "Geschirr", "Papier", "Schein", "Fuß", "Kapitän", "China", "Uhr", "Golf", "Nagel", "Bayern", "Schalter", "Pferd", "Kiel", "Feuer", "Riemen", "Punkt", "Kippe", "Geist", "Quelle", "Bau", "Reif", "Korb", "Schule", "Katze", "Gürtel", "Hahn", "Riegel", "Tag", "Hotel", "Jura", "Glocke", "Horst", "Drossel", "Anwalt", "Wolkenkratzer", "Verein", "Königin", "Bogen", "Skelett", "Mal", "Torte", "Dieb", "Roulette", "Osten", "Ägypten", "Zylinder", "Karte", "Figur", "London", "Muschel", "Luft", "Kater", "Doktor", "Strudel", "Hase", "Matte", "Bombe", "Winnetou", "Boot", "Laser", "Pirat", "Straße", "Botschaft", "Moos", "Deutschland", "Löffel", "Millionär", "Becken", "Peking", "Schokolade", "Schlange", "Schnee", "Einhorn", "Läufer", "Bank", "Kerze", "Amerikaner", "Schuppen", "Fessel", "Jahr", "Korn", "Tod", "Mond", "Landung", "Oper", "Tau", "Stadion", "Arm", "Wurm", "Knie", "Lakritze", "Fuchs", "Zug", "Taste", "Satellit", "Lippe", "Sekretär", "Tanz", "Genie", "Quartett", "Lager", "Gesicht", "Drache", "Afrika", "Rolle", "Öl", "Leben", "Ente", "Würfel", "Morgenstern", "Batterie", "Oktopus", "Jäger", "Zeit", "Kapelle", "Zitrone", "Leim", "Schale", "Blüte", "Geschoss", "Feder", "Saturn", "Po", "Theater", "Hexe", "Bund", "Schotten", "Glück", "Atlas", "Polizei", "Hamburger", "Alien", "Grund", "Auge", "Olymp", "Schild", "Schuh", "Schimmel", "Blau", "Ausdruck", "Gras", "Wirtschaft", "Taucher", "Iris", "Bergsteiger", "Elf", "Riese", "Mark", "Brause", "Himalaja", "Gehalt", "Zwerg", "Bande", "Krankenhaus", "Wurf", "Bindung", "Heide", "Nadel", "Kasino", "Mexiko"}
