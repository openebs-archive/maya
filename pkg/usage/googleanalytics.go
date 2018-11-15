package usage

import (
	"github.com/golang/glog"
	analytics "github.com/jpillora/go-ogle-analytics"
)

// Send sends a single usage metric to Google Analytics
func (u *Usage) Send(gaClient *analytics.Client) {
	go func() {
		event := analytics.NewEvent(u.category, u.action)
		event.Label(u.label)
		event.Value(u.value)
		if err := gaClient.Send(event); err != nil {
			glog.Errorf(err.Error())
			return
		}
	}()
}
