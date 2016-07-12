package main

import (
	"encoding/json"
	"fmt"
	"github.com/mpars0ns/sonar-es-go/helpers"
	"gopkg.in/olivere/elastic.v3"
	"log"
	//"os"
	"strings"
	"os"
)

type SonarImportedFile struct {
	File string `json:"file"`
	Sha1 string `json:"sha1"`
}

func main() {
	client, err := elastic.NewClient()
	import_check := helpers.Check_index_and_create("scansio-sonar-ssl-imported")
	if import_check == false {
		log.Fatal("We couldn't create a index properly exit out now!")
	}
	query := elastic.NewMatchAllQuery()
	searchResult, err := client.Search().Index("scansio-sonar-ssl-imported").Query(query).Do()
	if err != nil {
		log.Fatal("error running test MatchAll query on scansio index", err)
	}
	fmt.Println(searchResult.Hits.TotalHits)
	importedfiles := map[string]bool{}
	if searchResult.Hits.TotalHits > 0 {

		for _, hit := range searchResult.Hits.Hits {
			var t SonarImportedFile
			err := json.Unmarshal(*hit.Source, &t)

			if err != nil {
				fmt.Println("we had an error decoding hits from MatchAll: ", err)
			}
			importedfiles[t.File] = true
		}
	}
	res := helpers.DownloadFeed()
	for _, i := range res.Studies {
		if i.Uniqid == "sonar.ssl" {
			for _, f := range i.Files {
				fname := ""
				if strings.Contains(f.Name, "20131030-20150518") {
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
					checkdownload, _ := helpers.Check_downloaded(fname, f.Fingerprint)
					if checkdownload == true {
						check, _ := helpers.Check_sha1(fname, f.Fingerprint)
						if check == false {
							err := helpers.DownloadFile(f.Name, fname)
							if err != nil {
								log.Fatal("We had an error on downloading file ", fname, err)
							}
							fmt.Printf("Download of file %v is successful \n", fname)
						}
					}

					err := helpers.DownloadFile(f.Name, fname)
					if err != nil {
						log.Fatal("We had an error in downloading file ", fname, err)
					}
					fmt.Printf("Download of file %v is successful", fname)
					check, err := helpers.Check_sha1(fname, f.Fingerprint)
					if err != nil {
						fmt.Printf("Error with sha1 on this file %v with error %v \n", f.Name, err)
						continue
					}
					if check == false {
						fmt.Printf("Error with sha1 on this file %v \n", f.Name)
						continue
					}
					helpers.Process_Certs(fname)
					parsed_file := SonarImportedFile{File: fname, Sha1: f.Fingerprint}
					_, certserr := client.Index().Index("scansio-sonar-ssl-imported").Type("imported-file").Id(f.Fingerprint).BodyJson(parsed_file).Do()
					if certserr != nil {
						// Handle error
						panic(err)
					}
					removeCertsErr := os.Remove(fname)
					if removeCertsErr != nil {
						fmt.Println(removeCertsErr)
					}
				}

				if strings.Contains(fname, "hosts.gz") {
					fmt.Printf("We need to import %v \n", fname)
					fmt.Printf("%v %v %v %v \n", f.Name, f.Size, f.Fingerprint, f.UpdatedAt)
					checkdownload, _ := helpers.Check_downloaded(fname, f.Fingerprint)
					if checkdownload == true {
						check, _ := helpers.Check_sha1(fname, f.Fingerprint)
						if check == false {
							err := helpers.DownloadFile(f.Name, fname)
							if err != nil {
								log.Fatal("We had an error on downloading file ", fname, err)
							}
							fmt.Printf("Download of file %v is successful \n", fname)
						}
					} else {
						err := helpers.DownloadFile(f.Name, fname)
						if err != nil {
							log.Fatal("We had an error on downloading file ", fname, err)
						}
						fmt.Printf("Download of file %v is successful \n", fname)
						check, err := helpers.Check_sha1(fname, f.Fingerprint)
						if err != nil {
							fmt.Printf("Error with sha1 on this file %v with error %v \n", f.Name, err)
						}
						if check == false {
							fmt.Printf("Error with sha1 on this file %v \n", f.Name)
							continue
						}
					}
					helpers.Process_Hosts(fname)
					parsed_file := SonarImportedFile{File: fname, Sha1: f.Fingerprint}
					_, hostserr := client.Index().Index("scansio-sonar-ssl-imported").Type("imported-file").Id(f.Fingerprint).BodyJson(parsed_file).Do()
					if hostserr != nil {
						// Handle error
						panic(hostserr)
					}
					removeHostsErr := os.Remove(fname)
					if removeHostsErr != nil {
						fmt.Println(removeHostsErr)
					}
				}
			}
		}
	}
}
