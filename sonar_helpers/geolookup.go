package sonar_helpers

import (
	"fmt"
	"github.com/abh/geoip"
)

func lookup_ip(ip string) (c string, cc string, asn string, netblock string) {

	geofile := "geoip/GeoIP.dat"
	asnfile := "geoip/GeoIPASNum.dat"

	gi, err := geoip.Open(geofile)
	if err != nil {
		fmt.Printf("Unable to open GeoIP datanase \n")
	}
	giasn, err := geoip.Open(asnfile)
	if err != nil {
		fmt.Printf("Unable to open GeoIP ASN datanase \n")
	}

	if gi != nil {
		country, country_code := gi.GetCountry(ip)

	}

}
