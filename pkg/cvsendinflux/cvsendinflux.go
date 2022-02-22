package cvsendinflux

import (
	"regexp"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Data struct {
	Bucket      string
	Token       string
	InfluxUrl   string
	Measurement string
	Org         string
	Tags        map[string]string
	Fields      map[string]interface{}
}

func TrimTarget(Response string) string {
	Target := strings.Split(Response, ":")
	Trimmed := Target[4]
	//fmt.Println(Trimmed)
	//RemoveUpdate := strings.ReplaceAll(Trimmed, "} update", "")
	//ReturnTrimmed := strings.Trim(RemoveUpdate, "\"")
	re := regexp.MustCompile(`"[^"]+"`)
	newStrs := re.FindAllString(Trimmed, -1)
	var s string
	for _, s := range newStrs {
		return strings.Trim(s, "\"")
	}
	return strings.Trim(s, "\"")
}

func (c *Data) WriteInflux() {

	client := influxdb2.NewClient(c.InfluxUrl, c.Token)

	// always close clienfunc t at the end
	defer client.Close()

	// get non-blocking write client
	writeAPI := client.WriteAPI(c.Org, c.Bucket)

	p := influxdb2.NewPoint(c.Measurement,
		c.Tags,
		c.Fields,
		time.Now())
	// write point asynchronously
	writeAPI.WritePoint(p)
	writeAPI.Flush()
}
