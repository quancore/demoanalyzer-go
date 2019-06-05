package analyser

import (
	"math"

	common "github.com/quancore/demoanalyzer-go/common"
)

// ###### custom event structs ####
type eventCommon struct {
	Tick     int
	analyser *Analyser
}

type event interface {
	handleEvent()
}

// Scheduler schedule the user defined custom tasks
type Scheduler struct {
	// the next tick value a custom event will be handled
	earliestValidTick int
	tickRate          float64
	scheduledTasks    map[int][]event
}

// #####################

// NewScheduler instantiate new scheduler
func NewScheduler(tickRate float64) *Scheduler {
	scheduler := &Scheduler{tickRate: tickRate}
	scheduler.scheduledTasks = make(map[int][]event)
	return scheduler
}

func (sc Scheduler) addEvent(currentTick int, offsetSec float64, ev event) int {
	offsetTick := int(common.SecondsToTick(offsetSec, sc.tickRate))
	executionTick := currentTick + offsetTick
	if executionTick > 0 {
		sc.scheduledTasks[executionTick] = append(sc.scheduledTasks[executionTick], ev)
		if sc.earliestValidTick > executionTick {
			sc.earliestValidTick = executionTick
		}
	}
	return executionTick
}

func (sc Scheduler) checkEvent(tick int) {
	if tick >= sc.earliestValidTick {
		if eventList, ok := sc.scheduledTasks[tick]; ok {
			for _, currEvent := range eventList {
				currEvent.handleEvent()
			}
			delete(sc.scheduledTasks, tick)
			sc.earliestValidTick = sc.getMinKey(sc.scheduledTasks)
		}
	}
}

func (sc Scheduler) getMinKey(taskMap map[int][]event) int {
	minNumber := math.MaxInt32
	for k := range taskMap {
		if k < minNumber {
			minNumber = k
		}
	}

	return minNumber
}
