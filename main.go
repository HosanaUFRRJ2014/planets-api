package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
)

/*Planet ... Saves planet basic data*/
type Planet struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Climate string `json:"climate"`
	Terrain string `json:"terrain"`
}

//TODO: Remove these global vars.

var planetMetadata = table.Metadata{
	Name:    "planet",
	Columns: []string{"ID", "Name", "Climate", "Terrain"},
	SortKey: []string{"id"},
}

var planetTable = table.New(planetMetadata)

var session gocqlx.Session

// Database functions

func dbConnect() gocqlx.Session {
	cluster := gocql.NewCluster("192.168.1.1", "192.168.1.2", "192.168.1.3")
	cluster.Keyspace = "planet"
	cluster.Consistency = gocql.Quorum
	_session, error := gocqlx.WrapSession(
		cluster.CreateSession(),
	)

	if error != nil {
		log.Fatal(error)
		log.Fatal("Error while trying to connect to database")
	}
	return _session
}

func selectPlanetByParam(paramName string, paramValue ...interface{}) []Planet {
	if len(paramValue) != 1 {
		panic("Please, inform just one parameter for selectPlanetByParam")
	}
	var planets []Planet
	value := paramValue[0]

	stmt, stmtError := planetTable.Select()
	queryMap := qb.M{paramName: value}
	query := session.Query(stmt, stmtError).BindMap(queryMap)
	execError := query.SelectRelease(&planets)
	if execError != nil {
		log.Fatal(execError)
		log.Fatal("Error while retrieving planet by", paramName)
	}

	return planets
}

// **
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

func SearchByParam(paramName string, value ...interface{}) []Planet {
	// Acess method from db given name and returns Planet. Case insensitive
	var data []Planet
	data = selectPlanetByParam(paramName, value)

	//data = db.searchByName(name)

	// TODO: ADD try/catch para o caso de não encontrar o id
	// TODO: Add case insensitive search
	// TODO: Treat response in case its not found
	// for _, planet := range planets {
	// 	if planet.Name == name {
	// 		data = planet
	// 		break
	// 	}
	// }

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

func capitalizeName(name string) string {
	var capitalizedName string

	capitalizedName = strings.ToTitle(
		strings.ToLower(name),
	)

	return capitalizedName
}

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
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Fatal("Could not read body for request")
		panic(err)
	}

	var newPlanet Planet

	json.Unmarshal(body, &newPlanet)

	newPlanet.Name = capitalizeName(newPlanet.Name)

	created := addNewPlanet(newPlanet)

	// TODO: Return new id and new created object?
	response := map[string]bool{"created": created}

	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if created {
		writer.WriteHeader(http.StatusCreated)
	} else {
		writer.WriteHeader(http.StatusBadRequest)
	}
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
	param := "id"
	id := getByAttribute(param, request)
	idAsInt := parseIDToInt64(id)
	data := SearchByParam(param, idAsInt)

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	encoder.Encode(data)

}

func GetByName(writer http.ResponseWriter, request *http.Request) {
	param := "name"
	name := getByAttribute(param, request)
	capitalizedName := capitalizeName(name)
	data := SearchByParam(param, capitalizedName)

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
	router.HandleFunc("/planets", ListPlanets)
	router.HandleFunc("/planet", CreateNewPlanet).Methods("POST")
	router.HandleFunc("/planet/id/{id}", GetByID)
	router.HandleFunc("/planet/name/{name}", GetByName)
	router.HandleFunc("/planet/{id}", DeletePlanetByID).Methods("DELETE")

	log.Println("Listening at port 5555...")
	log.Fatal(http.ListenAndServe(":5555", router))
}

func main() {

	session = dbConnect()
	defer session.Close()

	planets = []Planet{
		Planet{ID: 0, Name: "Tatooine", Climate: "hot", Terrain: "sand"},
		Planet{ID: 1, Name: "Earth", Climate: "sunnie", Terrain: "rocks"},
	}

	//	planets := ListPlanets()

	//	fmt.Println(planets)

	handleRequests()
}
