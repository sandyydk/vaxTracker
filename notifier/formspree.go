package notifier

import (
	"fmt"
	"strings"
	"vaxtrack/config"
	"vaxtrack/utils"
)

const (
	formspreeHost = "https://formspree.io/f/{form_id}"
)

type Formspree struct {
	formID string
}

type FormspreeMessage struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func (f *Formspree) getMessagePayload(msg Message) FormspreeMessage {

	message := FormspreeMessage{
		Name:    "VaxTracker",
		Email:   "abc@gmail.com",
		Message: msg.Content,
	}

	return message
}

func (f *Formspree) Notify(message Message) error {

	url := strings.Replace(formspreeHost, "{form_id}", f.formID, 1)

	fmt.Printf("Notifying %v\n", url)
	req, err := utils.GetHTTPRequest("POST", url, f.getMessagePayload(message))
	if err != nil {
		return err
	}

	_, err = utils.ExecuteHTTPRequest(req)
	if err != nil {
		fmt.Printf("Error executing webhook call to formspree: %#v\n", err)
		return err
	}

	return nil
}

func NewFormspree(formID string) *Formspree {
	var f Formspree

	f.formID = formID

	return &f
}

func init() {
	config.InitializeConfig()
	// Access config fields here and load your object and register it into Notifiers
	f := NewFormspree(config.Conf.Notifier.Formspree.FormID)
	RegisterNotifier("formspree", f)
}
