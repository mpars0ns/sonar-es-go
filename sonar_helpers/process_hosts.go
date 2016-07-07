package sonar_helpers

import (
	"bufio"
	"compress/gzip"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"io"
	"log"
	"os"
	"time"
	"sync"
)

type Host struct {
	CountryCode string `json:"country_code"`
	City        string `json:"city"`
	Region      string `json:"region"`
	Asn         string `json:"asn"`
	Host        string `json:"host"`
	Hash        string `json:"hash"`
	Source      string `json:"source"`
	LastSeen    string `json:"last_seen"`
}

type ProcessGeoIP struct {
	id int
}


func file_reader(lookupchan chan Host, hostsfile string, wg sync.WaitGroup, Done chan struct{}) {

	defer wg.Done()

	fmt.Println(hostsfile)
	f, err := os.Open(hostsfile)
	if err != nil {
		log.Fatal("Error opening file, ", err)
	}
	defer f.Close()

	hf, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal("Error opening file, ", err)
	}
	defer hf.Close()

	reader := csv.NewReader(bufio.NewReader(hf))
	for {
		data, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				//stuff ehere for last bulk or something like that
				/*_, lasterr := bulkRequest.Do()
				if lasterr != nil {
					log.Println(err)
				} */
				break
			}
		}

		source := "sonar"
		host, hash := data[0], data[1]
		last_seen, _ := time.Parse("20060102", hostsfile[0:8])
		lastseen := last_seen.String()
		newhost := Host{Host: host, Hash: hash, LastSeen: lastseen, Source: source}
		lookupchan <- newhost

	}
}

func ESWriter(indexchan chan Host, wg sync.WaitGroup, Done chan struct{}){
	defer wg.Done()

	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	//let's check if index exists:
	exists, err := client.IndexExists("passive-ssl-sonar-hosts").Do()
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		_, err = client.CreateIndex("passive-ssl-sonar-hosts").Do()
		if err != nil {
			panic(err)
		}
	}
	p, bulkerr := client.BulkProcessor().Name("HostImporter").Workers(1).BulkActions(500).BulkSize(2 << 20).FlushInterval(30 * time.Second).Do()
	fmt.Println(bulkerr)
	if bulkerr != nil {
		fmt.Println(bulkerr)
	}
	for {
		select {
		case nh := <-indexchan:
			hasher := sha1.New()
			hash_string := nh.Host + nh.Hash + nh.Source
			hasher.Write([]byte(hash_string))
			id := hex.EncodeToString(hasher.Sum(nil))
			newho, _ := json.Marshal(nh)
			fmt.Printf("Uploading host with id %v, %v", id, string(newho))
			indexDoc := elastic.NewBulkUpdateRequest().Index("passive-ssl-sonar-hosts").Type("host").Id(id).Doc(string(newho)).DocAsUpsert(true)
			p.Add(indexDoc)
		case <-Done:
			break

		}
	}
	//bulkRequest := client.Bulk()

	elasticerr := p.Close()
	if elasticerr != nil {
		log.Println(err)
	}

}

func Process_Hosts(hostsfile string) {
	lookupchan := make(chan Host)
	indexchan := make(chan Host, 1000)
	var wg sync.WaitGroup
	Done := make(chan struct{})
	defer close(Done)

	for w := 1; w <= 3; w++ {
		go Lookup_ip(lookupchan, indexchan, Done)
	}

	go file_reader(lookupchan, hostsfile, wg, Done)
	go ESWriter(indexchan, wg, Done)

	wg.Wait()
}
