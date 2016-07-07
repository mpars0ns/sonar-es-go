#!/usr/bin/env bash

#wget http://geolite.maxmind.com/download/geoip/database/GeoLiteCountry/GeoIP.dat.gz
wget http://geolite.maxmind.com/download/geoip/database/GeoLiteCity.dat.gz
wget http://download.maxmind.com/download/geoip/database/asnum/GeoIPASNum.dat.gz
gunzip GeoLiteCity.dat.gz
mv GeoLiteCity.dat geoip/.
gunzip GeoIPASNum.dat.gz
mv GeoIPASNum.dat geoip/.