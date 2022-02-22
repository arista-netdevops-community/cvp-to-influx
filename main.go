package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aristanetworks/goarista/gnmi"

	cvinv "github.com/arista-netdevops-community/cvp-to-influx/pkg/cvinv"
	"github.com/arista-netdevops-community/cvp-to-influx/pkg/cvsendinflux"
	cvstream "github.com/arista-netdevops-community/cvp-to-influx/pkg/cvstream"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"gopkg.in/yaml.v2"
)

// Define the fields for yaml config.yaml file.
type config struct {
	Cvp_server  string `yaml:"cvp_server"`
	Path        string `yaml:"path"`
	CvpToken    string `yaml:"cvptoken"`
	InfluxToken string `yaml:"influxtoken"`
	Origin      string `yaml:"origin"`
	StreamMode  string `yaml:"streammode"`
	Measurement string `yaml:"measurement"`
	InfluxUrl   string `yaml:"influxurl"`
	Org         string `yaml:"influxorg"`
	Bucket      string `yaml:"influxbucket"`
}

func main() {
	yamlcfg := flag.String("config", "config.yaml", "Name of the config yaml file")
	flag.Parse()
	//Initialize C as the yaml place holder for static variables.
	var c config
	//open up config.yaml to hold the token and other variables.
	yamlFile, err := ioutil.ReadFile(*yamlcfg)

	if err != nil {
		log.Print("Cannot find yaml file", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	//Since the CVP server is set as x.x.x.x:8443 we need to remove :8443 in this case.
	splitPort := strings.Split(c.Cvp_server, ":")
	//Call the CvpData Method
	devs := cvinv.CvpData{Token: c.CvpToken,
		Url:    "/api/resources/inventory/v1/Device/all",
		Server: splitPort[0]}
	//Initialize a map for device inventory.
	CvInv := map[string]string{}
	//Append to the CvInv file
	for name, serial := range devs.CvpDevices(devs.Token, devs.Url, devs.Server) {
		CvInv[name] = serial
	}
	//Display the devices which are found on CVP with time stamp and logging.
	log.Print("Total Devices found from CVP")
	for name, serial := range CvInv {
		InfoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
		fmt.Printf(InfoLog.Prefix() + name + serial + time.Now().UTC().String())
		fmt.Println("/n")
	}
	time.Sleep(2 * time.Second)

	//Start with gNMI here

	//Initiate the gNMI struct with all the needed parameters
	g := &cvstream.GNMI_CFG{Addr: c.Cvp_server, Origin: c.Origin, Path: c.Path, StreamMode: c.StreamMode, Token: c.CvpToken}
	//Create a channel for gNMI data
	respChan := make(chan *pb.SubscribeResponse)
	for _, devs := range CvInv {
		g.CreateChan(devs, respChan)
	}
	// Initialize what Influx needs for tags and fields.
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	//Loop through the response data and print it.
	for Resp := range respChan {
		if Resp.GetUpdate() == nil {
		} else {
			for _, update := range Resp.GetUpdate().GetUpdate() {
				// Since we do not currently carry the target within the gNMI update currently in CVP response Update. had to mess around with string manipulation of the update.
				gnmitargets := gnmi.SplitPath(Resp.String())
				target := cvsendinflux.TrimTarget(gnmitargets[0])
				//Since the value can really be of any type need to match the value and switch based off of it.
				var value interface{}
				switch val := update.Val.Value.(type) {
				case *pb.TypedValue_UintVal:
					value = update.Val.GetUintVal()
				case *pb.TypedValue_AsciiVal:
					value = update.Val.GetAsciiVal()
				case *pb.TypedValue_BytesVal:
					value = update.Val.GetBytesVal()
				case *pb.TypedValue_DecimalVal:
					value = update.Val.GetDecimalVal()
				case *pb.TypedValue_FloatVal:
					value = update.Val.GetFloatVal()
				case *pb.TypedValue_StringVal:
					value = update.Val.GetStringVal()
				case nil:
					fmt.Println(val)

				}
				// Get the value of the Path element
				lastVal := update.Path.Elem[len(update.Path.Elem)-1]
				// Print the entire path with the value and such.
				log.Print("Path ", gnmi.StrPath(update.Path), " Target ", target, " PathValue ", value)
				for _, r := range update.Path.Elem {
					if len(r.Key) > 0 {
						tags["PathKey"] = r.Key["name"] // Ethernet9
						tags["Update"] = lastVal.Name   // in-octets
						tags["PathElem"] = r.Name       // interface
					}
					// Insert the Influx tags and values that are needed.
					tags["path"] = (gnmi.StrPath(update.Path)) //example interfaces/interface[name=Management0]/state/counters
					tags["origin"] = update.Path.Origin
					tags["target"] = target
					fields["PathValue"] = value // 16117243053
					insert := cvsendinflux.Data{
						Bucket:      c.Bucket,
						Token:       c.InfluxToken,
						Org:         c.Org,
						InfluxUrl:   c.InfluxUrl,
						Tags:        tags,
						Fields:      fields,
						Measurement: c.Measurement,
					}
					// If the value is a uint64 and not zero inset it.
					if value == update.Val.GetUintVal() && update.Val.GetUintVal() > 0 { // If a value is a uint64 and it is not zero then write to the DB.
						insert.WriteInflux()
					}
				}

			}
		}
	}
}
