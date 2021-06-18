package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
	"vaxtrack/config"
	"vaxtrack/utils"
)

const (
	host           = "https://cdn-api.co-vin.in/api"
	defaultTimeout = 5
)

func init() {

	config.InitializeConfig()
	// Access config fields here and load your object and register it into Notifiers
	f := NewAPISetu(config.Conf.Scheduler.DistrictIDs)
	RegisterScheduler("", f) // Register with the cron notation here that should be used or a default one will be used
}

type District struct {
	ID   int    `json:"district_id"`
	Name string `json:"district_name"`
}

type State struct {
	ID   int      `json:"state_id"`
	Name string   `json:"state_name"`
	Dist District `json:",omitempty"`
}

type StatesResponse struct {
	States []State `json:"states"`
}

type DistrictsResponse struct {
	Districts []District `json:"districts"`
}

type Center struct {
	CenterID     int    `json:"center_id"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	StateName    string `json:"state_name"`
	DistrictName string `json:"district_name"`
	BlockName    string `json:"block_name"`
	Pincode      int    `json:"pincode"`
	Lat          int    `json:"lat"`
	Long         int    `json:"long"`
	From         string `json:"from"`
	To           string `json:"to"`
	FeeType      string `json:"fee_type"`
	Sessions     []struct {
		SessionID         string   `json:"session_id"`
		Date              string   `json:"date"`
		AvailableCapacity int      `json:"available_capacity"`
		MinAgeLimit       int      `json:"min_age_limit"`
		Vaccine           string   `json:"vaccine"`
		Slots             []string `json:"slots"`
	} `json:"sessions"`
}

type CentersResponse struct {
	Centers []Center `json:"centers"`
}

//APISetu holds info of districts whose appointment availabilities are to be fetched
type APISetu struct {
	districtIDs  []int
	wg           sync.WaitGroup
	appointsChan chan Appointment
}

// NewAPISetu returns a handle which manages fetching appointments based on config provided
func NewAPISetu(districtIDs []int) *APISetu {
	return &APISetu{districtIDs: districtIDs}
}

func (s *APISetu) Initialize() error {

	s.appointsChan = make(chan Appointment)
	return nil
}

// ListStates lists out states and their IDs as per cowin
func (s *APISetu) ListStates() ([]State, error) {

	url := host + "/v2/admin/location/states"
	req, err := utils.GetHTTPRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating http request for listing states: %#v\n", err)
		return nil, err
	}

	resp, err := utils.ExecuteHTTPRequest(req)
	if err != nil {
		fmt.Printf("Error listing states from cowin: %#v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response bytes for %s: %v", url, err)
	}

	var statesResp StatesResponse

	err = json.Unmarshal(bodyBytes, &statesResp)
	if err != nil {
		fmt.Printf("Failed to unmarshal states response: %v\n", err)
		return nil, err
	}

	return statesResp.States, nil
}

// ListDistricts lists out districts under a state with their IDs as per cowin
func (s *APISetu) ListDistricts(stateID int) ([]District, error) {

	url := fmt.Sprintf("%s/v2/admin/location/districts/%d", host, stateID)
	req, err := utils.GetHTTPRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating http request for listing districts for state %d: %#v\n", stateID, err)
		return nil, err
	}

	resp, err := utils.ExecuteHTTPRequest(req)
	if err != nil {
		fmt.Printf("Error listing districts for state %d from cowin: %#v\n", stateID, err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response bytes for %s: %v", url, err)
	}

	var districtsResp DistrictsResponse

	err = json.Unmarshal(bodyBytes, &districtsResp)
	if err != nil {
		fmt.Printf("Failed to unmarshal districts response: %v\n", err)
		return nil, err
	}

	if districtsResp.Districts == nil || len(districtsResp.Districts) == 0 {
		return nil, fmt.Errorf("invalid state id provider: %d", stateID)
	}

	return districtsResp.Districts, nil
}

/// LOOK AT FAN-IN FAN-OUT for this. Send data to a chan of Appointments and merge all of them at one place and receive in one channel

// GetAppointmentsForADistrict fetches available appointments from the source for a given district
func (s *APISetu) GetAppointmentsForADistrict(ctx context.Context, districtID int) {
	defer s.wg.Done()
	//var appointments []Appointment

	today := time.Now().Format("02-01-2006")

	url := fmt.Sprintf("%s/v2/appointment/sessions/public/calendarByDistrict?district_id=%d&date=%s", host, districtID, today)
	fmt.Printf("URL: %+v\n", url)
	req, err := utils.GetHTTPRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating http request for listing appointments for districtID %d: %#v\n", districtID, err)
		return
	}

	resp, err := utils.ExecuteHTTPRequest(req)
	if err != nil {
		fmt.Printf("Error listing appointments for districtID %d from cowin: %#v\n", districtID, err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//return nil, fmt.Errorf("error reading response bytes for %s: %v", url, err)
		return
	}

	var centersResp CentersResponse

	//fmt.Printf("Body bytes: %#v\n", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &centersResp)
	if err != nil {
		fmt.Printf("Failed to unmarshal appointments response: %v\n", err)
		return
	}

	//fmt.Printf("Districts appointments response: %#v\n", centersResp)

	if centersResp.Centers == nil || len(centersResp.Centers) == 0 {
		//return nil, fmt.Errorf("no appointments found for %d", districtID)
		return
	}

	for _, center := range centersResp.Centers {

		for _, session := range center.Sessions {

			var appointment Appointment

			if session.AvailableCapacity <= 0 {
				continue
			}

			appointment.AvailableCapacity = int64(session.AvailableCapacity)
			appointment.Date = session.Date
			appointment.VaccineProvider = session.Vaccine
			appointment.Slots = session.Slots
			appointment.MinAgeLimit45 = session.MinAgeLimit == 45
			appointment.MinAgeLimit18 = session.MinAgeLimit == 18

			appointment.District = center.DistrictName
			appointment.From = center.From
			appointment.To = center.To
			appointment.Fee = center.FeeType
			appointment.Pincode = int64(center.Pincode)
			appointment.CenterName = center.Name
			appointment.CenterID = int64(center.CenterID)

			fmt.Printf("Writing to channel:%#v\n", appointment.CenterName)
			s.appointsChan <- appointment
			fmt.Printf("Wrote to channel\n")
		}

	}

	fmt.Printf("Exiting go routine\n")
}

// func (s *APISetu) MergeAppointments(appointments ...<-chan Appointment) <-chan Appointment {
// 	var wg sync.WaitGroup
// 	out := make(chan Appointment)

// 	output := func(a <-chan Appointment) {
// 		for n := range a {
// 			out <- n
// 		}
// 		wg.Done()
// 	}
// 	wg.Add(len(appointments))

// 	for _, c := range appointments {
// 		go output(c)
// 	}

// 	go func() {
// 		wg.Wait()
// 		close(out)
// 	}()

// 	return out
// }

func (s *APISetu) GetAppointments(ctx context.Context) ([]Appointment, error) {

	var appointments []Appointment

	fmt.Printf("get appointments for : %#v\n", s.districtIDs)

	for _, district := range s.districtIDs {
		s.wg.Add(1)
		fmt.Printf("Launching go routine for district: %#v\n", district)
		go s.GetAppointmentsForADistrict(ctx, district)
	}

	// Just to close the chan we launch this
	go func() {
		s.wg.Wait()
		fmt.Printf("Closing appointmentsChan\n")
		close(s.appointsChan)
	}()

	// To exit on context closure
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			}
		}
	}()

	// for n := range s.MergeAppointments(s.appointsChan) {

	// }
	fmt.Println("Ranging over appointsChan")

	for appointment := range s.appointsChan {
		fmt.Printf("Appointment appending: %#v\n", appointment.CenterName)
		appointments = append(appointments, appointment)
	}

	fmt.Printf("Finished sending. Now exiting\n")

	//	timer := time.NewTicker(time.Minute * defaultTimeout)
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return appointments, nil
	// 		//	case <-timer.C:

	// 	}
	// }

	return appointments, nil
}
