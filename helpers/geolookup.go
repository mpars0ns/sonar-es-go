package helpers

import (
	"fmt"
	"github.com/abh/geoip"
)

func Lookup_ip(lookup chan Host, index chan Host, Done chan struct{}) {
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
		case newhost := <-lookup:
			newhost.Asn = lookup_asn(newhost.Host, giasn)
			newhost.CountryCode, newhost.City, newhost.Region = lookup_geo(newhost.Host, gi)
			index <- newhost
		case <-Done:
			fmt.Println("Sending done from Lookup")
			return
		}

	}
}

/*
func Lookup_ip(ip string) (country string, city string, region string, asn string) {

        asn = lookup_asn(ip)
	country, city, region = lookup_geo(ip)

	return country, city, region, asn
}*/

func lookup_asn(ip string, giasn *geoip.GeoIP) (asn string) {
	asn = ""
	name, _ := giasn.GetName(ip)
	return name

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
