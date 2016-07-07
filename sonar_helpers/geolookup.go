package sonar_helpers

import (
	"fmt"
	"github.com/abh/geoip"
)

func Lookup_ip(lookup chan *Host, index chan *Host) {
	for {
		newhost := <-lookup
		newhost.Asn = lookup_asn(newhost.Host)
		newhost.CountryCode, newhost.City, newhost.Region = lookup_geo(newhost.Host)
		index <- newhost
	}
	close(lookup)

}

/*
func Lookup_ip(ip string) (country string, city string, region string, asn string) {

        asn = lookup_asn(ip)
	country, city, region = lookup_geo(ip)

	return country, city, region, asn
}*/

func lookup_asn(ip string) (asn string) {
	asn = ""
	asnfile := "geoip/GeoIPASNum.dat"
	giasn, err := geoip.Open(asnfile)
	if err != nil {
		fmt.Printf("Unable to open GeoIP ASN datanase \n")
	}
	if giasn != nil {
		name, _ := giasn.GetName(ip)
		return name
	}
	return asn

}

func lookup_geo(ip string) (country string, city string, region string) {
	country = ""
	city = ""
	region = ""
	geofile := "geoip/GeoLiteCity.dat"
	gi, err := geoip.Open(geofile)
	if err != nil {
		fmt.Printf("Unable to open GeoIP datanase \n")
	}

	if gi != nil {
		record := gi.GetRecord(ip)
		if record != nil {
			city = record.City
			country = record.CountryCode
			region = record.Region
			return country, city, region
		}

	}
	return country, city, region

}
