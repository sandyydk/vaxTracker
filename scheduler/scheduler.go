package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Scheduler checks for open slots for vaccination. Can write a scheduler for different sources like api setu or any other sites
type Scheduler interface {
	GetAppointments(context.Context) ([]Appointment, error)
	Initialize() error
}

var (
	once sync.Once
)

type Appointment struct {
	CenterID          int64    `json:"center_id"`
	CenterName        string   `json:"center_name"`
	CenterAddress     string   `json:"address"`
	State             string   `json:"state"`
	District          string   `json:"district"`
	Pincode           int64    `json:"pincode"`
	From              string   `json:"from"` // From time like 10:00
	To                string   `json:"to"`   // To time like 13:00
	Fee               string   `json:"vaccine_fee"`
	Date              string   `json:"date"`
	AvailableCapacity int64    `json:"availability_count"`
	MinAgeLimit45     bool     `json:"min_age_45"`
	MinAgeLimit18     bool     `json:"min_age_18"`
	ForAll            bool     `json:"for_all"`
	VaccineProvider   string   `json:"vaccine_type"`
	Slots             []string `json:"slots,omitempty"`
}

func (a *Appointment) String() string {
	return fmt.Sprintf(
		`An Appointment is available at the following center with the given details : - 
		Center Name: %s,
		Center Address: %s,
		Date: %v,
		From: %v,
		To: %v,
		Fee: %s,
		AvailabilityCount: %v,
		MinAgeLimitOf18: %v,
		MinAgeLimitOF45: %v,
		Slots: %v`,
		a.CenterName,
		a.CenterAddress,
		a.Date,
		a.From,
		a.To,
		a.Fee,
		a.AvailableCapacity,
		a.MinAgeLimit18,
		a.MinAgeLimit45,
		a.Slots)

}

// SchedulerMap holds list of supported schedulers who verify from different sources for possible vaccination slots
var SchedulerMap map[string]Scheduler

func RegisterScheduler(name string, scheduler Scheduler) {
	once.Do(
		func() {
			if SchedulerMap == nil {
				SchedulerMap = make(map[string]Scheduler)
			}
		})

	// Add to the map if not present
	if _, ok := SchedulerMap[strings.ToLower(name)]; !ok {
		SchedulerMap[strings.ToLower(name)] = scheduler
	}
}
