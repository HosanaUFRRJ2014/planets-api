package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

/*Planet ... Saves planet basic data*/
type Planet struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Climate string `json:"climate"`
	Terrain string `json:"terrain"`
}

//TODO: Remove this global var. Its here to simulate DB
var planets []Planet

func (planet Planet) movieAppearenceCount() int {
	// Get data from https://swapi.dev/api/planets
	// count movies appearence and return it

	return 0

}

func createPlanet(id int64, name string, climate string, terrain string) Planet {
	// Create planet of type Planet
	planet := Planet{ID: id, Name: name, Climate: climate, Terrain: terrain}

	// TODO: Adds it to database
	return planet
}

func getPlanets() []Planet {
	// Retrieve all planets from database and returns it
	var planets []Planet
	//planets := GetPlanetsFromDB()

	// Para cada planetadeve retornar também a quantidade de aparições em filmes

	return planets
}

func SearchByName(name string) Planet {
	// Acess method from db given name and returns Planet. Case insensitive
	var data Planet

	//data = db.searchByName(name)

	return data
}

func SearchByID(id int64) Planet {
	// Acess method from db given id and returns Planet.
	var data Planet

	//data = db.searchByID(id)

	// TODO: ADD try/catch para o caso de não encontrar o id
	data = planets[id]

	return data
}

func RemovePlanet(id int64) bool {
	// Removes a planet by id. If planet id not found, raises exception
	removed := false

	return removed
}

/*API functions*/

func APIHome(writer http.ResponseWriter, r *http.Request) {
	homeTemplate, _ := template.ParseFiles("home.html")
	homeTemplate.ExecuteTemplate(writer, "home.html", nil)
}

func ListPlanets(writer http.ResponseWriter, request *http.Request) {
	json.NewEncoder(writer).Encode(planets)
}

func GetByID(writer http.ResponseWriter, request *http.Request) {
	variables := mux.Vars(request)
	id := variables["id"]
	idAsInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Fatal("ID", id, "is not of type integer")
	}
	data := SearchByID(idAsInt)

	json.NewEncoder(writer).Encode(data)

}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", APIHome)
	router.HandleFunc("/planets", ListPlanets)
	router.HandleFunc("/planet/{id}", GetByID)

	log.Fatal(http.ListenAndServe(":5555", router))
}

func main() {

	planets = []Planet{
		Planet{ID: 0, Name: "Tatooine", Climate: "hot", Terrain: "sand"},
		Planet{ID: 1, Name: "Earth", Climate: "sunnie", Terrain: "plain"},
	}

	//	planets := ListPlanets()

	//	fmt.Println(planets)

	handleRequests()
}
