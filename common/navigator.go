package common

import (
	"fmt"
	"math"
	"os"
	"path"

	p_common "github.com/markus-wa/demoinfocs-golang/common"
	"github.com/mrazza/gonav"
	"github.com/quancore/demoanalyzer-go/utils"
	logging "github.com/sirupsen/logrus"
)

const (
	navMapDir = "nav_files"
)

type place struct {
	name           string
	area           float32
	center         gonav.Vector3
	navigatorPlace *gonav.NavPlace
	diameter       float32
	logger         *logging.Logger
}

// newPlace return a new place with attributes
func newPlace(currPlace *gonav.NavPlace, logger *logging.Logger) (*place, error) {
	var createdPlace place
	name := currPlace.Name
	center, err := currPlace.GetEstimatedCenter()
	if err == nil {
		createdPlace = place{name: name, center: center, navigatorPlace: currPlace, logger: logger}
		createdPlace.area, createdPlace.diameter = createdPlace.calculatePlaceAttributes()
	}

	return &createdPlace, err
}

// calculatePlaceAttributes calculate place attributes such as area, place occupancy diameter etc.
func (currPlace *place) calculatePlaceAttributes() (float32, float32) {
	var squareMeter float32
	var northZ, southZ float32
	var changedNorth, changedSouth bool
	var mostWestX, mostSouthY = float32(math.MaxFloat32), float32(math.MaxFloat32)
	var mostNorthY, mostEastX = float32(-math.MaxFloat32), float32(-math.MaxFloat32)
	for _, area := range currPlace.navigatorPlace.Areas {
		// set area of the place
		squareMeter += area.GetRoughSquaredArea()
		// get a place furthest points in the direction of east, west, north, south using included areas
		SW, NE := area.GetSouthWestPoint(), area.GetNorthEastPoint()
		areaWestX, areaEastX, areaNorthY, areaSouthY := SW.X, NE.X, NE.Y, SW.Y
		mostWestX, _ = utils.Minf32(mostWestX, areaWestX)
		mostEastX, _ = utils.Maxf32(mostEastX, areaEastX)
		// if the furthest point changed, record the Z value of that point as well
		mostSouthY, changedSouth = utils.Minf32(mostSouthY, areaSouthY)
		if changedSouth {
			southZ = SW.Z
		}
		mostNorthY, changedNorth = utils.Maxf32(mostNorthY, areaNorthY)
		if changedNorth {
			northZ = NE.Z
		}

		currPlace.logger.WithFields(logging.Fields{
			"place name": currPlace.name,
			"area name":  area.String(),
			"south west": SW,
			"north east": NE,
		}).Debug("Area iteration for place attribute calculation")
	}

	// calculate diameter for the distance of the place from center
	// to use place assignment

	// get place with and height of the place using found furthest points
	placeWidth := utils.Absf32(mostEastX - mostWestX)
	placeHeight := utils.Absf32(mostNorthY - mostSouthY)

	// calculate cosinus of the angle between nort and south points in between y and z axis
	northPoint := gonav.Vector3{X: 0, Y: mostNorthY, Z: northZ}
	southPoint := gonav.Vector3{X: 0, Y: mostSouthY, Z: southZ}
	northPoint.Sub(southPoint)
	hipotenus := northPoint.Length()
	cosAngle := placeHeight / hipotenus

	// calculate min diameter for in 2D dimension (x, y). We set diameter the half
	// of dimensions
	diameterX, diameterY := (placeWidth / 2.0), (placeHeight / 2.0)
	diameter, _ := utils.Minf32(diameterX, diameterY)

	// reflect this diamater to third dimension (z)
	reflectedDiameter := diameter
	if cosAngle != 0 || !math.IsNaN(float64(cosAngle)) {
		reflectedDiameter = diameter / cosAngle
	}

	currPlace.logger.WithFields(logging.Fields{
		"place name":          currPlace.name,
		"most west x":         mostWestX,
		"most east x":         mostEastX,
		"most south y":        mostSouthY,
		"most north y":        mostNorthY,
		"north z":             northZ,
		"south z":             southZ,
		"diameter 2d":         diameter,
		"cosine of the slope": cosAngle,
		"diameter 3d":         reflectedDiameter,
	}).Debug("Place attribute calculation")

	return squareMeter, reflectedDiameter
}

// placeAssignment place assignment struct to store single place assignment
type placeAssignment struct {
	tick       int
	controller p_common.Team
	place      *place
}

// Navigator struct of navigator object to store all place assignments
type Navigator struct {
	log                *logging.Logger
	mesh               gonav.NavMesh
	totalMapArea       float32
	placeMap           map[string]*placeAssignment
	tAssignedPlaceMap  map[string]*placeAssignment
	ctAssignedPlaceMap map[string]*placeAssignment
	lastPlayerPlaces   map[int64]string //player id - place name map
	// tick value mainly used determining whether a player has moved or not
	previousTickAssign int
}

// NewNavigator return new navigator object
func NewNavigator(log *logging.Logger) *Navigator {
	newNavigator := Navigator{log: log}
	newNavigator.placeMap = make(map[string]*placeAssignment)
	newNavigator.tAssignedPlaceMap = make(map[string]*placeAssignment)
	newNavigator.ctAssignedPlaceMap = make(map[string]*placeAssignment)
	newNavigator.lastPlayerPlaces = make(map[int64]string)

	newNavigator.previousTickAssign = -1
	return &newNavigator
}

// ResetNavigator reset state of the navigator object
func (navigator *Navigator) ResetNavigator() {
	navigator.log.Info("Reseting navigator state")
	for _, place := range navigator.placeMap {
		place.controller = p_common.TeamUnassigned
		place.tick = -1
	}
	navigator.previousTickAssign = -1
	navigator.tAssignedPlaceMap = make(map[string]*placeAssignment)
	navigator.ctAssignedPlaceMap = make(map[string]*placeAssignment)
	navigator.lastPlayerPlaces = make(map[int64]string)

}

// Parse parse nav file of given map
func (navigator *Navigator) Parse(mapName string) error {
	var err error
	navMeshName := fmt.Sprintf("%s.nav", mapName)
	gopath := utils.GetGoPath()
	mainDirPath := "src/github.com/quancore/demoanalyzer-go"
	fullPath := path.Join(gopath, mainDirPath, navMapDir, navMeshName)

	f, openErr := os.Open(fullPath) // Open the file
	if readErr, ok := openErr.(*os.PathError); ok {
		navigator.log.WithFields(logging.Fields{
			"error": readErr.Error(),
		}).Error("File failed to read ")
		return readErr
	}
	parser := gonav.Parser{Reader: f}
	navigator.mesh, err = parser.Parse() // Parse the file
	if err != nil {
		return err
	}
	navigator.log.Info("Mesh has succesfully parsed")

	places := navigator.mesh.Places

	for _, place := range places {
		currPlace, err := newPlace(place, navigator.log)
		if err == nil {
			initAssignment := placeAssignment{tick: -1, controller: p_common.TeamUnassigned, place: currPlace}
			navigator.placeMap[place.Name] = &initAssignment
			navigator.totalMapArea += currPlace.area
			navigator.log.WithFields(logging.Fields{
				"name":         place.Name,
				"square meter": currPlace.area,
				"center":       currPlace.center,
				"diameter":     currPlace.diameter,
			}).Debug("Place has been initilized")
		}

	}

	return nil
}

// GetTotalMapArea get total map area
func (navigator *Navigator) GetTotalMapArea() float32 { return navigator.totalMapArea }

// AssignPlace assing nearest place to a given player
func (navigator *Navigator) AssignPlace(player *PPlayer, tick int) {
	// get player teamstate
	if player.TeamState == nil {
		navigator.log.WithFields(logging.Fields{
			"tick":        tick,
			"player name": player.Name,
			"event name":  "map assignment",
		}).Error("Player state is nil")
		return
	}
	playerTeamState := player.TeamState
	pos := player.Position
	posNav := gonav.Vector3{X: float32(pos.X), Y: float32(pos.Y), Z: float32(pos.Z)}
	nearestArea := navigator.mesh.QuadTreeAreas.FindAreaByPoint(posNav, true)
	if nearestArea != nil {
		nearestPlace := nearestArea.Place
		if nearestPlace != nil {
			if currentAssignment, ok := navigator.placeMap[nearestPlace.Name]; ok {
				navigator.makeAssignment(player.GetSteamID(), playerTeamState, posNav, nearestPlace, currentAssignment, tick)
			} else {
				navigator.log.WithFields(logging.Fields{
					"tick":        tick,
					"player name": player.Name,
					"place name":  nearestPlace.Name,
				}).Error("The place has not been initilized")
			}
		} else {
			navigator.log.WithFields(logging.Fields{
				"tick":        tick,
				"player name": player.Name,
				"area name":   nearestArea.String(),
			}).Debug("The area has not been assigned to any place")
		}

	} else {
		navigator.log.WithFields(logging.Fields{
			"tick":        tick,
			"player name": player.Name,
		}).Debug("Player is not near any area")
	}
}

func (navigator *Navigator) makeAssignment(playerID int64, playerTeamState *p_common.TeamState, posNav gonav.Vector3, nearestPlace *gonav.NavPlace, currAssignment *placeAssignment, tick int) {
	var assigmentDone bool
	currAssignedTeam := currAssignment.controller
	placeName := nearestPlace.Name
	playerTeam := playerTeamState.Team()

	navigator.log.WithFields(logging.Fields{
		"tick":          tick,
		"place name":    placeName,
		"assigned team": currAssignedTeam,
		"player team":   playerTeam,
	}).Debug("Checking assignment for a player")

	// calculate distance to center
	placeCenter := currAssignment.place.center
	maxDistance := currAssignment.place.diameter
	placeCenter.Sub(posNav)
	distance := placeCenter.Length()
	if distance < maxDistance {

		// make place assignment to the player
		assigmentDone = navigator.makeAssignmentByCase(playerID, playerTeamState, placeName, currAssignment, tick)

	} else {
		navigator.log.WithFields(logging.Fields{
			"tick":               tick,
			"place name":         placeName,
			"player team":        playerTeam,
			"current controller": currAssignment.controller,
			"place center":       placeCenter,
			"player postion":     posNav,
			"place diameter":     maxDistance,
			"player distance":    distance,
		}).Debug("Player out of distance to any place")
	}

	if assigmentDone == true {
		navigator.log.WithFields(logging.Fields{
			"tick":        tick,
			"place name":  placeName,
			"old team":    currAssignedTeam,
			"new team":    currAssignment.controller,
			"player team": playerTeam,
		}).Debug("Place assignment has been made")
	}

}

func (navigator *Navigator) makeAssignmentByCase(playerID int64, playerTeamState *p_common.TeamState, placeName string, currAssignment *placeAssignment, tick int) bool {
	var assigmentDone bool

	currAssignedTeam := currAssignment.controller
	assignedTick := currAssignment.tick
	playerTeam := playerTeamState.Team()

	// update player last place
	navigator.lastPlayerPlaces[playerID] = placeName

	// if already assigned to the team
	if currAssignedTeam == playerTeam {
		currAssignment.tick = tick
		navigator.log.WithFields(logging.Fields{
			"tick":       tick,
			"place name": placeName,
			"old team":   currAssignedTeam,
			"new team":   currAssignment.controller,
		}).Debug("The place has already assigned to same team")
		return assigmentDone
	}

	// if the place has already been assigned to opponent
	if currAssignedTeam == playerTeamState.Opponent.Team() {
		// if the place has been assigned at the same time, two opponent members
		// still in the area meaning place should be unassigned
		if tick == assignedTick {
			navigator.log.WithFields(logging.Fields{
				"tick":          tick,
				"assigned tick": assignedTick,
				"place name":    placeName,
				"old team":      currAssignedTeam,
				"player team":   playerTeam,
			}).Debug("Two opponent in the same place. Place will be unassigned")
			currAssignment.controller = p_common.TeamUnassigned
			currAssignment.tick = tick

			navigator.removeTeamAssignment(currAssignedTeam, placeName)
			assigmentDone = true

		} else if tick > assignedTick { // if the place assigned before we assign the place to new team
			currAssignment.controller = playerTeam
			currAssignment.tick = tick
			navigator.addTeamAssignment(playerTeam, currAssignment)

			assigmentDone = true
			navigator.log.WithFields(logging.Fields{
				"tick":          tick,
				"assigned tick": assignedTick,
				"place name":    placeName,
				"old team":      currAssignedTeam,
				"player team":   playerTeam,
			}).Debug("The assigned place assigned to a new team")

		} else {
			navigator.log.WithFields(logging.Fields{
				"tick":        tick,
				"place name":  placeName,
				"player team": playerTeam,
				"old team":    currAssignedTeam,
				"new team":    currAssignment.controller,
			}).Info("bad place")
		}
	} else if currAssignedTeam == p_common.TeamUnassigned {
		if tick > assignedTick {
			currAssignment.controller = playerTeam
			currAssignment.tick = tick

			navigator.addTeamAssignment(playerTeam, currAssignment)

			assigmentDone = true
			navigator.log.WithFields(logging.Fields{
				"tick":          tick,
				"assigned tick": assignedTick,
				"place name":    placeName,
				"old team":      currAssignedTeam,
				"player team":   playerTeam,
			}).Debug("Unassigned place assigned to a new team")
		} else {
			navigator.log.WithFields(logging.Fields{
				"tick":          tick,
				"assigned tick": assignedTick,
				"place name":    placeName,
				"old team":      currAssignedTeam,
				"player team":   playerTeam,
			}).Debug("The place is shared with two teams and it remains unassigned for this check period")
		}

	}

	return assigmentDone
}

// getTeamAssignment get assignment maps of teams
func (navigator *Navigator) getTeamAssignment(playerTeam p_common.Team) (map[string]*placeAssignment, bool) {
	if playerTeam == p_common.TeamTerrorists {
		return navigator.tAssignedPlaceMap, true
	} else if playerTeam == p_common.TeamCounterTerrorists {
		return navigator.ctAssignedPlaceMap, true
	}

	return nil, false
}

// addTeamAssignment add a place to team map assignment
func (navigator *Navigator) addTeamAssignment(playerTeam p_common.Team, currAssignment *placeAssignment) {
	if teamPlaceMap, ok := navigator.getTeamAssignment(playerTeam); ok {
		teamPlaceMap[currAssignment.place.name] = currAssignment
	}
}

// removeTeamAssignment remove a place from team map assignment
func (navigator *Navigator) removeTeamAssignment(playerTeam p_common.Team, currPlaceName string) {
	if teamPlaceMap, ok := navigator.getTeamAssignment(playerTeam); ok {
		delete(teamPlaceMap, currPlaceName)
	}
}

// UpdateTeamAssignment update tick value of place - team assignments
// mostly used for stationary players
func (navigator *Navigator) UpdateTeamAssignment(playerID int64, playerTeamState *p_common.TeamState, tick int) {
	if placeName, ok := navigator.lastPlayerPlaces[playerID]; ok {
		playerTeam := playerTeamState.Team()
		if teamPlaceMap, ok := navigator.getTeamAssignment(playerTeam); ok {
			for assignedPlaceName, placeAssigned := range teamPlaceMap {
				if placeName == assignedPlaceName {
					navigator.makeAssignmentByCase(playerID, playerTeamState, placeName, placeAssigned, tick)
				}
			}
		}
	}
}

// GetLastAssignedTick get the tick of last map assignment checking
func (navigator *Navigator) GetLastAssignedTick() int { return navigator.previousTickAssign }

// SetLastAssignedTick set the tick of last map assignment checking
func (navigator *Navigator) SetLastAssignedTick(previousTickAssign int) {
	navigator.previousTickAssign = previousTickAssign
}

// CalculateMapOccupancy calculate occupied area for two teams
func (navigator *Navigator) CalculateMapOccupancy() (float32, float32) {
	tArea := navigator.calculateTeamMapOccupancy(p_common.TeamTerrorists)
	ctArea := navigator.calculateTeamMapOccupancy(p_common.TeamCounterTerrorists)

	navigator.log.WithFields(logging.Fields{
		"total place number":       len(navigator.placeMap),
		"t assigned place number":  len(navigator.tAssignedPlaceMap),
		"ct assigned place number": len(navigator.ctAssignedPlaceMap),
	}).Debug("Place assignment summary")

	return tArea, ctArea
}

// calculateTeamMapOccupancy calculate suare meter occupancy for single team
func (navigator *Navigator) calculateTeamMapOccupancy(playerTeam p_common.Team) float32 {
	var totalArea float32
	if teamPlaceMap, ok := navigator.getTeamAssignment(playerTeam); ok {
		for _, placeAssigned := range teamPlaceMap {
			totalArea += placeAssigned.place.area
		}
	}

	return totalArea
}
