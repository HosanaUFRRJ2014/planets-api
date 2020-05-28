package main

import (
	"fmt"

	"github.com/google/uuid"
)

/*Planet ... Saves planet basic data*/
type Planet struct {
	id      uuid.UUID
	name    string
	climate string //TODO: change type
	terrain string //TODO: change type
}

func (planet Planet) movieAppearenceCount() int {
	// Get data from https://swapi.dev/api/planets
	// count movies appearence and return it

	return 0

}

func createPlanet(name string, climate string, terrain string) Planet {
	// Create planet of type Planet
	id := uuid.New()
	planet := Planet{id: id, name: name, climate: climate, terrain: terrain}

	// TODO: Adds it to database
	return planet
}

func ListPlanets() []Planet {
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

func SearchByID(id uuid.UUID) Planet {
	// Acess method from db given id and returns Planet.
	var data Planet

	//data = db.searchByID(id)

	return data
}

func RemovePlanet(id uuid.UUID) bool {
	// Removes a planet by id. If planet id not found, raises exception
	removed := false

	return removed
}

func main() {

	createPlanet("Tatooine", "sunnie", "acid")

	planets := ListPlanets()

	fmt.Println(planets)
}
