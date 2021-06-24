package usage

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Send the analytic results to server
// if wait is not set then Send will execute go routine
func (u *Usage) Send(wait bool) error {
	var fn func() error

	if u.url == "" {
		fn = u.SendGoogleAnalytic
	} else {
		fn = u.SendToUrl
	}

	if wait {
		return fn()
	}

	go fn()
	return nil
}

func (u *Usage) SendToUrl() error {
	val := url.Values{}

	u.setValues(val)

	str := val.Encode()
	buf := bytes.NewBufferString(str)

	resp, err := http.DefaultClient.Post(u.url, "application/x-www-form-urlencoded", buf)
	if err != nil {
		return err
	}

	fmt.Printf("got status %+v\n", resp)
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return errors.Errorf("Failed to send data, response code=%d", resp.StatusCode)
	}

	return nil
}

func (u *Usage) setValues(val url.Values) {
	if u.campaignSource != "" {
		val.Add("cs", u.campaignSource)
	}

	if u.campaignName != "" {
		val.Add("cn", u.campaignName)
	}

	if u.clientID != "" {
		val.Add("cid", u.clientID)
	}

	if u.appID != "" {
		val.Add("aid", u.appID)
	}

	if u.appVersion != "" {
		val.Add("av", u.appVersion)
	}

	if u.dataSource != "" {
		val.Add("ds", u.dataSource)
	}

	if u.appName != "" {
		val.Add("an", u.appName)
	}

	if u.appInstallerID != "" {
		val.Add("aiid", u.appInstallerID)
	}

	if u.documentTitle != "" {
		val.Add("dt", u.documentTitle)
	}

	if u.label != "" {
		val.Add("el", u.label)
		val.Add("ev", fmt.Sprintf("%d", u.value))
	}

	if u.Gclient.trackID != "" {
		val.Add("tid", u.Gclient.trackID)
	}
	return
}
