package helpers

import (
	"bufio"
	"compress/gzip"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"io"
	"log"
	"os"
	"time"

	"encoding/json"
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
	FirstSeen   string `json:"first_seen,omitempty"`
	Id          string `json:"id,omitempty"`
}

func (h *Host) SetFirstSeen(ts string) {
	h.FirstSeen = ts
}

func file_reader(lookupchan *chan Host, hostsfile string, Done chan struct{}) {

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
				return
			}
		}
		if len(data) < 2 {
			continue
		}
		source := "sonar"
		host, hash := data[0], data[1]
		lastSeen, _ := time.Parse("20060102", hostsfile[0:8])
		lastseen := lastSeen.Format(time.RFC3339)
		newhost := Host{}
		if hostsfile[0:8] == "20131030" {
			firstseen := lastseen
			newhost.FirstSeen = firstseen
			newhost.LastSeen = lastseen
			newhost.Host = host
			newhost.Hash = hash
			newhost.Source = source
		} else {
			newhost.LastSeen = lastseen
			newhost.Host = host
			newhost.Hash = hash
			newhost.Source = source
		}
		select {
		case *lookupchan <- newhost:
		case <-Done:
			return
		}
	}
}

func ESWriter(indexchan *chan Host, esWg *sync.WaitGroup, Done chan struct{}) {
	defer esWg.Done()
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal("error connecting to ES", err)
	}

	p, bulkerr := client.BulkProcessor().Name("HostImporter").Workers(1).BulkActions(1000).BulkSize(2 << 20).FlushInterval(30 * time.Second).Do()
	if bulkerr != nil {
		fmt.Println("Problem with the elastic bulk importer", bulkerr)
	}
OuterLoop:
	for {
		select {
		case nh := <-*indexchan:
			hasher := sha1.New()
			hash_string := nh.Host + nh.Hash + nh.Source
			hasher.Write([]byte(hash_string))
			id := hex.EncodeToString(hasher.Sum(nil))
			indexDoc := elastic.NewBulkUpdateRequest().Index("passive-ssl-sonar-hosts").Type("host").Id(id).Doc(nh).DocAsUpsert(true)
			p.Add(indexDoc)
		case <-Done:
			break OuterLoop
		}
	}
	flusherr := p.Flush()
	if flusherr != nil {
		log.Println("Error in final flush", flusherr)
	}
	elasticerr := p.Close()
	if elasticerr != nil {
		log.Println("Error in closing bulk processor", err)
	}
}

func search_newhosts() {
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal("error connecting to ES", err)
	}
	p, bulkerr := client.BulkProcessor().Name("HostImporter").Workers(1).BulkActions(500).BulkSize(2 << 20).FlushInterval(30 * time.Second).Do()
	if bulkerr != nil {
		fmt.Println("Problem with the elastic bulk importer", bulkerr)
	}
	query := elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewExistsQuery("first_seen"))
	sr, err := client.Scan().Index("passive-ssl-sonar-hosts").Query(query).FetchSource(true).Do()
	if err != nil {
		log.Fatal(err)
	}

	if sr.TotalHits() > 0 {
		fmt.Printf("Found a total of %d hosts\n", sr.TotalHits())
		for {
			res, err := sr.Next()
			if err == elastic.EOS {
				fmt.Println("EOS END Do one last query and see what happens")
				p.Flush()
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			// Iterate through results
			for _, hit := range res.Hits.Hits {
				var t Host
				id := hit.Id
				err := json.Unmarshal(*hit.Source, &t)
				if err != nil {
					log.Fatal(err)
				}
				t.SetFirstSeen(t.LastSeen)
				indexDoc := elastic.NewBulkUpdateRequest().Index("passive-ssl-sonar-hosts").Type("host").Id(id).Doc(t).DocAsUpsert(true)
				p.Add(indexDoc)
			}
		}
	} else {
		fmt.Print("Found no hosts\n")
	}
	p.Flush()

}

func Process_Hosts(hostsfile string) {
	checkCreateSonaHostsSSLIndex()
	lookupchan := make(chan Host)
	indexchan := make(chan Host)
	lookupDone := make(chan struct{})
	Done := make(chan struct{})
	var esWg sync.WaitGroup
	var lookupWg sync.WaitGroup

	lookupWg.Add(4)
	for w := 1; w <= 4; w++ {
		go Lookup_ip(&lookupchan, &indexchan, &lookupWg, lookupDone)
	}
	esWg.Add(2)
	for ew := 1; ew <= 2; ew++ {
		go ESWriter(&indexchan, &esWg, Done)
	}

	fmt.Println("Starting import at: ", time.Now())
	file_reader(&lookupchan, hostsfile, Done)
	close(lookupDone)
	lookupWg.Wait()
	close(Done)
	esWg.Wait()
	fmt.Println("Finished import at: ", time.Now())
	fmt.Println("Update first_seen started at: ", time.Now())
	// Now we need to go back and update...hopefully it works
	search_newhosts()
	search_newhosts() // run a second time to see if we find any stragglers
	fmt.Println("Update first_seen finished at: ", time.Now())

}
