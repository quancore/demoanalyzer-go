package analyser

import (
	"math"

	common "github.com/quancore/demoanalyzer-go/common"
	logging "github.com/sirupsen/logrus"
)

// ###### custom event structs ####
type eventCommon struct {
	offsetSec  float64
	isPeriodic bool
	// final tick value a periodic task will be scheduled
	endTick  int
	analyser *Analyser
}

type event interface {
	handleEvent(tick int)
	reschedule(tick int, scheduler *Scheduler, ev event)
	postEventHandler()
}

// Scheduler schedule the user defined custom tasks
type Scheduler struct {
	// the next tick value a custom event will be handled
	earliestValidTick int
	tickRate          float64
	scheduledTasks    map[int][]event
	analyser          *Analyser
}

func (ec eventCommon) reschedule(tick int, scheduler *Scheduler, ev event) {
	if ec.isPeriodic == true {
		// if we still need to schedule
		if tick < ec.endTick {
			scheduler.addEvent(tick, ec.offsetSec, ev)
		} else {
			// after finishing periodic tasks, we call post event handler
			ev.postEventHandler()
		}

	}
}

// #####################

// NewScheduler instantiate new scheduler
func NewScheduler(analyser *Analyser, tickRate float64) *Scheduler {
	scheduler := &Scheduler{tickRate: tickRate, analyser: analyser, earliestValidTick: math.MaxInt32}
	scheduler.scheduledTasks = make(map[int][]event)
	return scheduler
}

func (sc *Scheduler) addEvent(currentTick int, offsetSec float64, ev event) int {
	offsetTick := int(common.SecondsToTick(offsetSec, sc.tickRate))
	executionTick := currentTick + offsetTick
	if executionTick > 0 {
		sc.scheduledTasks[executionTick] = append(sc.scheduledTasks[executionTick], ev)
		if sc.earliestValidTick > executionTick {
			sc.earliestValidTick = executionTick
		}
	}

	sc.analyser.log.WithFields(logging.Fields{
		"tick":          currentTick,
		"offsetSec":     offsetSec,
		"executionTick": executionTick,
		"offset tick":   offsetTick,
		"earliest tick": sc.earliestValidTick,
	}).Debug("Event addition")
	return executionTick
}

func (sc *Scheduler) checkEvent(tick int) {
	if tick >= sc.earliestValidTick {
		sc.analyser.log.WithFields(logging.Fields{
			"tick":                tick,
			"earliest event tick": sc.earliestValidTick,
		}).Info("Check event called")
		if eventList, ok := sc.scheduledTasks[tick]; ok {
			// current round is ongoing, the event is already valid
			if sc.analyser.checkRoundEventValid(tick) {
				for _, currEvent := range eventList {
					currEvent.handleEvent(tick)
					currEvent.reschedule(tick, sc, currEvent)
				}
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
