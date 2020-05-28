package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
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

func addNewPlanet(newPlanet Planet) bool {
	created := false

	// TODO: Add it do database
	planets = append(planets, newPlanet)

	return created
}

func getAllPlanets() []Planet {
	// Retrieve all planets from database and returns it
	//var planets []Planet
	//planets := GetPlanetsFromDB()

	// Para cada planetadeve retornar também a quantidade de aparições em filmes

	return planets
}

func SearchByName(name string) Planet {
	// Acess method from db given name and returns Planet. Case insensitive
	var data Planet

	//data = db.searchByName(name)

	// TODO: ADD try/catch para o caso de não encontrar o id
	// TODO: Add case insensitive search
	// TODO: Treat response in case its not found
	for _, planet := range planets {
		if planet.Name == name {
			data = planet
			break
		}
	}

	return data
}

func SearchByID(id int64) Planet {
	// Acess method from db given id and returns Planet.
	var data Planet

	//data = db.searchByID(id)

	// TODO: ADD try/catch para o caso de não encontrar o id
	// TODO: Treat response in case its not found
	data = planets[id]

	return data
}

func RemovePlanet(id int64) bool {
	// Removes a planet by id. If planet id not found, raises exception
	//var planet Planet
	removed := false

	// TODO: Remove planet in DB
	//returns it?

	return removed
}

/* API Utils*/
func getByAttribute(attributeName string, request *http.Request) string {
	variables := mux.Vars(request)
	return variables[attributeName]
}

func parseIDToInt64(id string) int64 {
	idAsInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Fatal("ID", id, "is not of type integer")
	}

	return idAsInt
}

/*API functions*/

func APIHome(writer http.ResponseWriter, request *http.Request) {
	homeTemplate, _ := template.ParseFiles("home.html")
	homeTemplate.ExecuteTemplate(writer, "home.html", nil)
}

func CreateNewPlanet(writer http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)

	var newPlanet Planet

	json.Unmarshal(body, &newPlanet)

	created := addNewPlanet(newPlanet)

	// TODO: Return new id and new created object?
	response := map[string]bool{"created": created}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(response)
}

func ListPlanets(writer http.ResponseWriter, request *http.Request) {
	planets := getAllPlanets()

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(planets)
}

func GetByID(writer http.ResponseWriter, request *http.Request) {
	id := getByAttribute("id", request)
	idAsInt := parseIDToInt64(id)
	data := SearchByID(idAsInt)

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(data)

}

func GetByName(writer http.ResponseWriter, request *http.Request) {
	name := getByAttribute("name", request)
	data := SearchByName(name)

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(data)
}

func DeletePlanetByID(writer http.ResponseWriter, request *http.Request) {
	id := getByAttribute("id", request)
	idAsInt := parseIDToInt64(id)

	deleted := RemovePlanet(idAsInt)

	response := map[string]bool{"deleted": deleted}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(response)
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", APIHome)
	router.HandleFunc("/planet", CreateNewPlanet).Methods("POST")
	router.HandleFunc("/planets", ListPlanets)
	router.HandleFunc("/planet/id/{id}", GetByID)
	router.HandleFunc("/planet/name/{name}", GetByName)
	router.HandleFunc("/planet/delete/{id}", DeletePlanetByID)

	log.Fatal(http.ListenAndServe(":5555", router))
}

func main() {

	planets = []Planet{
		Planet{ID: 0, Name: "Tatooine", Climate: "hot", Terrain: "sand"},
		Planet{ID: 1, Name: "Earth", Climate: "sunnie", Terrain: "rocks"},
	}

	//	planets := ListPlanets()

	//	fmt.Println(planets)

	handleRequests()
}
