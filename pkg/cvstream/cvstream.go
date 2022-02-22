package cvstream

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"github.com/aristanetworks/goarista/gnmi"
	"golang.org/x/sync/errgroup"
)

type GNMI_CFG struct {
	Token      string
	Url        string
	Server     string
	Origin     string
	Path       string
	Mode       string
	StreamMode string
	Addr       string
}

func (c *GNMI_CFG) CreateChan(target string, channel chan *pb.SubscribeResponse) {
	InfoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	fmt.Printf(InfoLog.Prefix() + "Started streaming using target ID   " + target + time.Now().UTC().String())
	log.Print("\n")
	var cfg = &gnmi.Config{
		Addr:  c.Addr,
		Token: c.Token,
	}

	paths := []string{c.Path}
	ctx := gnmi.NewContext(context.Background(), cfg)
	client, err := gnmi.Dial(cfg)
	if err != nil {
		log.Print("Cannot Dial to CVP through gNMI interface ", err)
	}

	var subscribeOptions = &gnmi.SubscribeOptions{}
	subscribeOptions.Origin = c.Origin
	subscribeOptions.Target = target
	subscribeOptions.StreamMode = c.StreamMode
	subscribeOptions.Paths = gnmi.SplitPaths(paths)

	var g errgroup.Group
	g.Go(func() error {
		return gnmi.SubscribeErr(ctx, client, subscribeOptions, channel)
	})
}
