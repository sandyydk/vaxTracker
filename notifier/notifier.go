package notifier

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

var once sync.Once

// Notifier notifes through different channels like slack, formspree,etc. Can be extended to support more
type Notifier interface {
	Notify(msg Message) error
}

type Message struct {
	// Meta    interface{}
	Content string
	Time    string
}

type NotificationManager struct {
	notifyMsg chan Message
}

func NewManager() *NotificationManager {
	return &NotificationManager{
		notifyMsg: make(chan Message, 100),
	}
}

func (m *NotificationManager) Start(ctx context.Context) {
	var err error
	for {
		select {
		case content, ok := <-m.notifyMsg:
			if !ok {
				// channel has been closed
				return
			}

			for notifierName, notifier := range NotifierMap {
				err = notifier.Notify(content)
				if err != nil {
					fmt.Printf("Error notifying the notifying source %s: %#v\n", notifierName, err)
					break
				}
			}

		case <-ctx.Done():
			return

		}
	}
}

// NotifierMap holds list of supported notifiers to whom notifications would be sent
var NotifierMap map[string]Notifier

func RegisterNotifier(name string, notifier Notifier) {
	once.Do(
		func() {
			if NotifierMap == nil {
				NotifierMap = make(map[string]Notifier)
			}
		})

	// Add to the map if not present
	if _, ok := NotifierMap[strings.ToLower(name)]; !ok {
		NotifierMap[strings.ToLower(name)] = notifier
	}
}

func (m *NotificationManager) HandleNotifications(message string) {
	msgToNotify := Message{
		Content: message,
		Time:    time.Now().Format(time.RFC3339),
	}

	m.notifyMsg <- msgToNotify

}
