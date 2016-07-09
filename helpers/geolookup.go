package helpers

import (
	"fmt"
	"github.com/abh/geoip"
	"sync"
)

func Lookup_ip(lookup *chan Host, index *chan Host, lookupWg *sync.WaitGroup, Done chan struct{}) {
	defer lookupWg.Done()

	giasn, err := geoip.Open("geoip/GeoIPASNum.dat")
	if err != nil {
		fmt.Printf("Unable to open GeoIP ASN Database")
	}
	geofile := "geoip/GeoLiteCity.dat"
	gi, err := geoip.Open(geofile)
	if err != nil {
		fmt.Printf("Unable to open GeoIP datanase \n")
	}
	for {
		select {
		case newhost := <-*lookup:
			newhost.Asn = lookup_asn(newhost.Host, giasn)
			newhost.CountryCode, newhost.City, newhost.Region = lookup_geo(newhost.Host, gi)
			*index <- newhost
		case <-Done:
			return
		}

	}

}

func lookup_asn(ip string, giasn *geoip.GeoIP) (asn string) {
	asnname, _ := giasn.GetName(ip)
	return asnname

}

func lookup_geo(ip string, gi *geoip.GeoIP) (country string, city string, region string) {
	country = ""
	city = ""
	region = ""

	record := gi.GetRecord(ip)
	if record != nil {
		city = record.City
		country = record.CountryCode
		region = record.Region
		return country, city, region
	}

	return country, city, region

}
