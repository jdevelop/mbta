package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/jdevelop/mbta/gtfs-realtime"
	"net/http"
	"io/ioutil"
	"github.com/twpayne/go-kml"
	"io"
	"github.com/julienschmidt/httprouter"
	"flag"
	"fmt"
	"strings"
)

func buildKML(lineFilter string, writer io.Writer) {

	resp, _ := http.Get("http://developer.mbta.com/lib/GTRTFS/Alerts/VehiclePositions.pb")

	msg := gtfs_realtime.FeedMessage{}

	data, _ := ioutil.ReadAll(resp.Body)

	proto.Unmarshal(data, &msg)

	folders := map[string]*kml.CompoundElement{}

	d := kml.Document()
	k := kml.KML(d)

	for _, ent := range msg.Entity {
		routeName := *ent.Vehicle.Trip.RouteId
		if !strings.HasPrefix(routeName, lineFilter) {
			continue
		}
		f := folders[routeName]
		if f == nil {
			f = kml.Folder(kml.Name(routeName))
			folders[routeName] = f
			d.Add(f)
		}
		f.Add(
			kml.Placemark(
				kml.Name(*ent.Vehicle.Vehicle.Label),
				kml.Point(
					kml.Coordinates(
						kml.Coordinate{
							Lon: float64(*ent.Vehicle.Position.Longitude),
							Lat: float64(*ent.Vehicle.Position.Latitude),
						},
					),
				),
			),
		)
	}

	k.WriteIndent(writer, "", "  ")

}

func main() {
	p := flag.Int("port", 4000, "Port number")
	flag.Parse()

	svc := httprouter.New()
	svc.GET("/kml", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/vnd.google-earth.kml+xml")
		buildKML("CR-", w)
	})

	fmt.Printf("Listening on port %1d\n", *p)

	http.ListenAndServe(fmt.Sprintf(":%1d", *p), svc)

}
