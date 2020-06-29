package api


import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/HosanaUFRRJ2014/planets-api/model"
	"github.com/HosanaUFRRJ2014/planets-api/planet"
)

/*Structure of a response from SWAPI planets endpoint */
type SWAPIResponse struct {
	Next string `json:"next"`
	Planets [] SWAPIPlanet `json:"results"`
}

/*Structure of planet from SWAPI that matters*/
type SWAPIPlanet struct {
	Name    string `json:"name"`
	URL string `json:"url"`
	Films [] string `json:"films"`
}

const PLANETS_SWAPI_URL string = "https://swapi.dev/api/planets/"

/* API Utils*/

/*Applies trim by space and capitalization*/
func prepareString(name string) string {
	trimmedName := strings.Trim(name, " ")
	capitalizedName := capitalizeName(trimmedName)

	return capitalizedName
}

func capitalizeName(name string) string {
	var capitalizedName string

	if len(name) >= 1 {
		capitalizedName = strings.ToUpper(name[0:1]) + strings.ToLower(name[1:])
	} else {
		capitalizedName = strings.ToUpper(name)
	}

	return capitalizedName
}

func getByAttribute(attributeName string, request *http.Request) string {
	variables := mux.Vars(request)
	return variables[attributeName]
}

func parseIDToInt64(id string) int64 {
	idAsInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Fatal("ID ", id, " is not of type integer")
	}

	return idAsInt
}

func formatResponse(writer *http.ResponseWriter, data ...interface{}) {
	encoder := json.NewEncoder(*writer)
	encoder.SetIndent("", "\t")

	if data[0] == nil {
		//removes nil from response
		data = data[:len(data)-1] 
	}
	encoder.Encode(data)
}

func formatPlanetResponse(writer *http.ResponseWriter, planet model.Planet) {
	if planet.IsEmpty() {
		formatResponse(writer, nil)
	} else {
		formatResponse(writer, planet)
	}
}

func getSWAPIResponse(url, planetName string) SWAPIResponse {

	searchParam := "?search=" + planetName
	url = url + searchParam
	response, err := http.Get(url)

	if err != nil {
		log.Println(err.Error())
		log.Println("Error: Could not connect to SWAPI!")
	}

	responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err.Error())
	}
	
	var responseObject SWAPIResponse
	json.Unmarshal(responseData, &responseObject)

	return responseObject
}

/*API functions*/

func getAppearencesCountFromSWAPI(planetName string) (int, string) {
	var appearencesCount int = 0
	var planetURL string = ""
	
	responseObject := getSWAPIResponse(PLANETS_SWAPI_URL, planetName)

	swapiPlanets := responseObject.Planets

	if len(swapiPlanets) > 0 {
		swapiPlanet := swapiPlanets[0]
		films := swapiPlanet.Films
		if len(films) > 0 {
			appearencesCount = len(films)
		}
		planetURL = swapiPlanet.URL

	}

	return appearencesCount, planetURL
}

func APIHome(writer http.ResponseWriter, request *http.Request) {
	homeTemplate, _ := template.ParseFiles("home.html", "static/css/apistyle.css")
	homeTemplate.ExecuteTemplate(writer, "home.html", nil)
}

func CreateNewPlanet(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Fatal("Could not read body for request")
		panic(err)
	}

	var newPlanet model.Planet
	json.Unmarshal(body, &newPlanet)

	newPlanet.Name = prepareString(newPlanet.Name)
	// Get appearences count
	appearencesCount, planetSwapiURL := getAppearencesCountFromSWAPI(newPlanet.Name)

	//Updates new planet with swapi information
	newPlanet.AppearencesCount = appearencesCount
	newPlanet.PlanetSwapiURL = planetSwapiURL

	// Saving Planet
	created := planet.AddNewPlanet(newPlanet)

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
	planets := planet.GetAllPlanets()

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	if len(planets) == 0 {
		empty := []int{}
		encoder.Encode(empty)
	} else {
		encoder.Encode(planets)
	}

}

func GetByID(writer http.ResponseWriter, request *http.Request) {
	param := "id"
	id := getByAttribute(param, request)

	planet := planet.SearchByParam(param, id)

	formatPlanetResponse(&writer, planet)

}

func GetByName(writer http.ResponseWriter, request *http.Request) {
	param := "name"
	name := getByAttribute(param, request)
	capitalizedName := capitalizeName(name)
	planet := planet.SearchByParam(param, capitalizedName)

	formatPlanetResponse(&writer, planet)
}

func DeletePlanetByName(writer http.ResponseWriter, request *http.Request) {
	paramName := "name"
	name := getByAttribute(paramName, request)
	capitalizedName := capitalizeName(name)

	deleted := planet.RemovePlanetByParam(paramName, capitalizedName)
	response := map[string]bool{"deleted": deleted}
	formatResponse(&writer, response)
}

func DeletePlanetByID(writer http.ResponseWriter, request *http.Request) {
	paramName := "id"
	id := getByAttribute(paramName, request)

	deleted := planet.RemovePlanetByParam(paramName, id)
	response := map[string]bool{"deleted": deleted}
	formatResponse(&writer, response)
}

func HandleRequests(port string) {
	var dir string
	flag.StringVar(&dir, ".", "static/", "")
    flag.Parse()
	
	// Create the router
	router := mux.NewRouter().StrictSlash(true)
	
	// Serve static Files under http://localhost:{PORT}/static/<filename>
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	// Add routers
	router.HandleFunc("/", APIHome)
	router.HandleFunc("/planets", ListPlanets).Methods("GET")
	router.HandleFunc("/planet", CreateNewPlanet).Methods("POST")
	router.HandleFunc("/planet/id/{id}", GetByID).Methods("GET")
	router.HandleFunc("/planet/name/{name}", GetByName).Methods("GET")
	router.HandleFunc("/planet/id/{id}", DeletePlanetByID).Methods("DELETE")
	router.HandleFunc("/planet/name/{name}", DeletePlanetByName).Methods("DELETE")

	
	address := "127.0.0.1:" + port
	service := &http.Server{
		Handler:      router,
        Addr:         address,
        // Enforce timeouts for server
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
	}
	
	log.Println("Listening at port", port, "...")
	log.Fatal(service.ListenAndServe())
}