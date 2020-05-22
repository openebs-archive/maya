//go:generate go run generate/protocol.go

package ga

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
)

var trackingIDMatcher = regexp.MustCompile(`^UA-\d+-\d+$`)

func NewClient(trackingID string) (*Client, error) {
	if !trackingIDMatcher.MatchString(trackingID) {
		return nil, fmt.Errorf("Invalid Tracking ID: %s", trackingID)
	}
	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error){
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, network, "8.8.8.8:53")
			},
		},
	}
	return &Client{
		UseTLS:             true,
		HttpClient:         &http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
		},
		protocolVersion:    "1",
		protocolVersionSet: true,
		trackingID:         trackingID,
		clientID:           "go-ga",
		clientIDSet:        true,
	}, nil
}

type hitType interface {
	addFields(url.Values) error
}

func (c *Client) Send(h hitType) error {

	cpy := c.Copy()

	v := url.Values{}

	cpy.setType(h)

	err := cpy.addFields(v)
	if err != nil {
		return err
	}

	err = h.addFields(v)
	if err != nil {
		return err
	}

	gaUrl := ""
	if cpy.UseTLS {
		gaUrl = "https://www.google-analytics.com/collect"
	} else {
		gaUrl = "http://ssl.google-analytics.com/collect"
	}

	str := v.Encode()
	buf := bytes.NewBufferString(str)

	resp, err := c.HttpClient.Post(gaUrl, "application/x-www-form-urlencoded", buf)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Rejected by Google with code %d", resp.StatusCode)
	}

	// fmt.Printf("POST %s => %d\n", str, resp.StatusCode)

	return nil
}
