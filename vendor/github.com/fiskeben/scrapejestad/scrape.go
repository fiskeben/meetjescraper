package scrapejestad

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

// Read downloads a document and parses it.
func Read(u *url.URL) ([]Reading, error) {
	return ReadWithContext(context.Background(), u)
}

// ReadWithContext downloads a document and parses it.
func ReadWithContext(ctx context.Context, u *url.URL) ([]Reading, error) {
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	client := http.Client{Timeout: time.Second * 2}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error reading '%s': %v", u.String(), err)
	}

	var doc []JsonReading
	d := json.NewDecoder(res.Body)
	if err := d.Decode(&doc); err != nil {
		return nil, fmt.Errorf("error unmarshaling data: '%v'", err)
	}

	return mapJsonReadingsToReadings(doc)
}

func mapJsonReadingsToReadings(r []JsonReading) ([]Reading, error) {
	res := make([]Reading, len(r))
	for i, doc := range r {
		t, err := time.Parse("2006-01-02 15:04:05", doc.Timestamp)
		if err != nil {
			return nil, err
		}

		res[i] = Reading{
			SensorID: strconv.Itoa(doc.Id),
			Time:     t.Unix(),
			Date:     t,
			Temp:     doc.Temperature,
			Humidity: doc.Humidity,
			Voltage:  doc.Supply,
			Firmware: strconv.Itoa(doc.FirmwareVersion),
			Position: Position{
				Lat: doc.Latitude,
				Lng: doc.Longitude,
			},
		}
	}

	return res, nil
}

func parse(r io.Reader) ([]Reading, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return parseSubtree(doc)
}

func parseSubtree(n *html.Node) ([]Reading, error) {
	if n.Type == html.ElementNode && n.Data == "table" {
		return parseTable(n.FirstChild.NextSibling)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res, err := parseSubtree(c)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, nil
}

func parseTable(t *html.Node) ([]Reading, error) {
	rows := make([]Reading, 0, 10)
	for c := t.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "tr" {
			continue
		}
		nodes := mapRow(c)
		switch len(nodes) {
		case 0:
			continue
		case 5:
			g, err := parseGateway(nodes)
			if err != nil {
				fmt.Printf("error parsing gateway: %v\n", err)
				continue
			}
			row := rows[len(rows)-1]
			row.Gateways = append(row.Gateways, g)
			rows[len(rows)-1] = row
		case 17:
			row, err := parseRow(nodes)
			if err != nil {
				fmt.Printf("error parsing row: %v\n", err)
				continue
			}
			rows = append(rows, *row)
		default:
			fmt.Printf("node %v has unexpected number of nodes: %d\n", c.Data, len(nodes))
		}
	}
	return rows, nil
}

func parseRow(n []*html.Node) (*Reading, error) {
	var r Reading

	r.SensorID = getID(n[0])

	data := strings.TrimSpace(n[1].FirstChild.Data)
	t, err := time.Parse("2006-01-02 15:04:05", data)
	if err != nil {
		return nil, err
	}
	r.Date = t
	r.Time = t.Unix()

	data = strings.TrimSpace(n[2].FirstChild.Data)
	v := data[:len(data)-3]
	temp, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return nil, err
	}
	r.Temp = float32(temp)

	data = strings.TrimSpace(n[3].FirstChild.Data)
	v = data[:len(data)-1]
	h, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return nil, err
	}
	r.Humidity = float32(h)

	data = strings.TrimSpace(n[7].FirstChild.Data)
	v = data[:len(data)-1]
	p, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return nil, err
	}
	r.Voltage = float32(p)

	r.Firmware = strings.TrimSpace(n[9].FirstChild.Data)

	pos, err := parsePosition(n[10])
	if err != nil {
		return nil, err
	}
	r.Position = pos

	data = strings.TrimSpace(n[11].FirstChild.Data)
	fcnt, err := strconv.Atoi(data)
	if err != nil {
		return nil, err
	}
	r.Fcnt = fcnt

	g, err := parseGateway(n[12:])
	if err != nil {
		return nil, err
	}
	r.Gateways = []Gateway{g}

	return &r, nil
}

func parseGateway(n []*html.Node) (Gateway, error) {
	var g Gateway

	parent := n[0].FirstChild
	if parent.FirstChild != nil {
		pos, _ := extractPositionFromURL(parent)
		g.Position = pos
		g.Name = strings.TrimSpace(parent.FirstChild.Data)
	}

	data := strings.TrimSpace(n[1].FirstChild.Data)
	if len(data) > 2 {
		v := data[:len(data)-2]
		dist, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return g, err
		}
		g.Distance = float32(dist)
	}

	rssi, err := strconv.ParseFloat(strings.TrimSpace(n[2].FirstChild.Data), 32)
	if err != nil {
		return g, err
	}
	g.RSSI = float32(rssi)

	lsnr, err := strconv.ParseFloat(strings.TrimSpace(n[3].FirstChild.Data), 32)
	if err != nil {
		return g, err
	}
	g.LSNR = float32(lsnr)

	parts := strings.Split(strings.TrimSpace(n[4].FirstChild.Data), ",")
	freq, err := strconv.ParseFloat(parts[0][:len(parts[0])-3], 32)
	if err != nil {
		return g, err
	}
	s := RadioSettings{
		Frequency: float32(freq),
		Sf:        strings.TrimSpace(parts[1]),
		Cr:        strings.TrimSpace(parts[2]),
	}
	g.RadioSettings = s

	return g, nil
}

func mapRow(n *html.Node) []*html.Node {
	if n.FirstChild.Data == "th" {
		return make([]*html.Node, 0)
	}
	res := make([]*html.Node, 0, 5)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Data != "td" {
			continue
		}
		res = append(res, c)
	}
	return res
}

func getID(n *html.Node) string {
	c := n.FirstChild
	if id := strings.TrimSpace(c.Data); id != "" {
		return id
	}
	c = c.NextSibling
	if id := strings.TrimSpace(c.Data); id != "" && id != "a" {
		return id
	}
	c = c.FirstChild
	if id := strings.TrimSpace(c.Data); id != "" && id != "a" {
		return id
	}
	return ""
}

func parsePosition(n *html.Node) (Position, error) {
	if n == nil || n.FirstChild == nil || n.FirstChild.NextSibling == nil {
		return Position{}, nil
	}

	data := n.FirstChild.NextSibling.FirstChild.Data
	parts := strings.Split(strings.TrimSpace(data), " ")
	lat, err := strconv.ParseFloat(parts[0], 32)
	if err != nil {
		return Position{}, err
	}
	lng, err := strconv.ParseFloat(parts[len(parts)-1], 32)
	if err != nil {
		return Position{}, err
	}
	return Position{Lat: float32(lat), Lng: float32(lng)}, nil
}

func extractPositionFromURL(n *html.Node) (Position, error) {
	var uri string
	var pos Position

	for _, a := range n.Attr {
		if a.Key == "href" {
			uri = a.Val
			break
		}
	}
	if uri == "" {
		return pos, nil
	}

	lat, err := extractPositionPart(uri, "mlat")
	if err != nil {
		return pos, err
	}
	pos.Lat = lat

	lng, err := extractPositionPart(uri, "mlon")
	if err != nil {
		return pos, err
	}
	pos.Lng = lng

	return pos, nil
}

func extractPositionPart(uri, name string) (float32, error) {
	i := strings.Index(uri, name)
	if i == -1 {
		return 0, nil
	}

	s := uri[i+5:]
	endAt := strings.Index(s, "&")
	if endAt == -1 {
		endAt = len(s)
	}
	s = s[:endAt]
	val, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, fmt.Errorf("unable to parse latitude from '%s': %v", uri, err)
	}

	return float32(val), nil
}
