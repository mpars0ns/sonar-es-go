package main

import (
	"encoding/json"
	"fmt"
	"github.com/mpars0ns/sonar-es-go/sonar_helpers"
	"gopkg.in/olivere/elastic.v3"
	"log"
	"strings"
)

func main() {
	client, err := elastic.NewClient()
	import_check := sonar_helpers.Check_index_and_create("scansio-sonar-ssl-imported")
	if import_check == false {
		log.Fatal("We couldn't create a index properly exit out now!")
	}
	query := elastic.NewMatchAllQuery()
	searchResult, err := client.Search().Index("scansio-sonar-ssl-imported").Query(query).Do()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(searchResult.Hits.TotalHits)
	importedfiles := map[string]bool{}
	if searchResult.Hits.TotalHits > 0 {
		type SonarImportedFile struct {
			File string `json:"file"`
			Sha1 string `json:"sha1"`
		}
		for _, hit := range searchResult.Hits.Hits {
			var t SonarImportedFile
			err := json.Unmarshal(*hit.Source, &t)

			if err != nil {
				fmt.Println("we had an error: ", err)
			}
			importedfiles[t.File] = true
		}
	}
	res := sonar_helpers.DownloadFeed()
	for _, i := range res.Studies {
		if i.Uniqid == "sonar.ssl" {
			for _, f := range i.Files {
				fname := ""
				if strings.Contains(f.Name, "20131030-20150518") {
					fmt.Println("We have the granddaddy")
					fname = "20131030-20150518_certs.gz"
				} else {
					fname = f.Name[48:65]
				}
				if importedfiles[fname] {
					fmt.Printf("Already imported %v", fname)
					continue
				}
				if strings.Contains(fname, "certs.gz") {
					fmt.Printf("We need to import %v \n", fname)
					fmt.Printf("%v %v %v %v \n", f.Name, f.Size, f.Fingerprint, f.UpdatedAt)
					err := sonar_helpers.DownloadFile(f.Name, fname)
					if err != nil {
						log.Fatal("We had an error in downloading file ", fname, err)
					}
					fmt.Printf("Download of file %v is successful %v", fname)
					check, err := sonar_helpers.Check_sha1(fname, f.Fingerprint)
					if err != nil {
						fmt.Printf("Error with sha1 on this file %v with error \n", f.Name, err)
						continue
					}
					if check == false {
						fmt.Printf("Error with sha1 on this file %v \n", f.Name)
						continue
					}
				}
				if strings.Contains(fname, "hosts.gz") {
					fmt.Printf("We need to import %v \n", fname)
					fmt.Printf("%v %v %v %v \n", f.Name, f.Size, f.Fingerprint, f.UpdatedAt)
					checkdownload, _ := sonar_helpers.Check_downloaded(fname, f.Fingerprint)
					if checkdownload == true {
						check, _ := sonar_helpers.Check_sha1(fname, f.Fingerprint)
						if check == false {
							err := sonar_helpers.DownloadFile(f.Name, fname)
							if err != nil {
								log.Fatal("We had an error on downloading file ", fname, err)
							}
							fmt.Printf("Download of file %v is successful %v \n", fname)
						}
					} else {
						err := sonar_helpers.DownloadFile(f.Name, fname)
						if err != nil {
							log.Fatal("We had an error on downloading file ", fname, err)
						}
						fmt.Printf("Download of file %v is successful %v \n", fname)
						check, err := sonar_helpers.Check_sha1(fname, f.Fingerprint)
						if err != nil {
							fmt.Printf("Error with sha1 on this file %v with error \n", f.Name, err)

						}
						if check == false {
							fmt.Printf("Error with sha1 on this file %v \n", f.Name)
							continue
						}

					}
					sonar_helpers.Process_Hosts(fname)
					//country, city, region, asn := sonar_helpers.Lookup_ip("8.8.8.8")
					//fmt.Printf("We got Country: %v\n City: %v\n Region: %v\n ASN: %v\n", country, city, region, asn)

				}

			}
		}

	}

}
