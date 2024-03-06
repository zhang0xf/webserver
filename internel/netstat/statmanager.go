package netstat

import (
	"fmt"
	"time"
)

var isRunning int32
var manager *StatManager

type StatManager struct {
	startChan chan *NetStat
	stopChan  chan *NetStat
	stats     map[*NetStat]bool
	start     chan struct{}
}

func newStatManager() *StatManager {
	return &StatManager{
		startChan: make(chan *NetStat, 1000),
		stopChan:  make(chan *NetStat, 1000),
		stats:     make(map[*NetStat]bool),
		start:     make(chan struct{}),
	}
}

func (statManager *StatManager) run() {
	close(statManager.start)
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case s := <-statManager.startChan:
			fmt.Println("add a conn's stat")
			statManager.stats[s] = true
		case s := <-statManager.stopChan:
			fmt.Println("remove a conn's stat")
			delete(statManager.stats, s)
		case <-ticker.C:
			fmt.Println("statManager calc all stats")
			for stat := range statManager.stats {
				stat.doCalc()
			}
		}
	}
}
