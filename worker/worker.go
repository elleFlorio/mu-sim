package worker

import (
	"log"
	"math/rand"
	"time"
)

const (
	c_MAXITER = 10000
	c_LOW     = 1000
	c_MEDIUM  = 5000
	c_HEAVY   = 10000
)

var (
	source rand.Source
	gen    *rand.Rand
)

func init() {
	source = rand.NewSource(time.Now().UnixNano())
	gen = rand.New(source)
}

func Work(workload string) {
	var load float64

	switch workload {
	case "none":
		return
	case "low":
		load = gen.ExpFloat64() * c_LOW
	case "medium":
		load = gen.ExpFloat64() * c_MEDIUM
	case "heavy":
		load = gen.ExpFloat64() * c_HEAVY
	default:
		log.Println("Undefined workload ", workload)
		return
	}

	timer := time.NewTimer(time.Millisecond * time.Duration(load))
	for {
		select {
		case <-timer.C:
			return
		default:
			cpuTest()
		}
	}
}

func cpuTest() float64 {
	plusMinus := false
	pi := 0.0
	for i := 1.0; i < c_MAXITER; i = i + 2.0 {
		if plusMinus {
			pi -= 4.0 / i
			plusMinus = false
		} else {
			pi += 4.0 / i
			plusMinus = true
		}
	}
	return pi
}
