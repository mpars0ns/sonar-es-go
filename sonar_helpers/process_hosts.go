package sonar_helpers

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
	"sync"
	"time"

	"encoding/json"
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
				fmt.Println("Got EOF we should be done!!!")
				return

			}
		}

		source := "sonar"
		host, hash := data[0], data[1]
		last_seen, _ := time.Parse("20060102", hostsfile[0:8])
		lastseen := last_seen.Format(time.RFC3339)
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
		case lookupchan <- newhost:
 		case <- Done: return
 		}

	}
}

func ESWriter(indexchan chan Host, wg sync.WaitGroup, Done chan struct{}) {
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
		mapping := `{
    "settings":{
        "number_of_shards":5,
        "number_of_replicas":0
    },
    "mappings":{
         "host" : {
        "properties" : {
          "host": {"type": "ip", "index": "analyzed"},
          "hash": {"type": "string"},
          "first_seen": {"type": "date", "format": "dateOptionalTime"},
          "last_seen": {"type": "date", "format": "dateOptionalTime"},
          "asn": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "country_code": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "city": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "region": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "port": {"type": "integer"},
          "source": {"type": "string"}
        }
      }
        }
    }
}`
		_, err = client.CreateIndex("passive-ssl-sonar-hosts").BodyString(mapping).Do()
		if err != nil {
			panic(err)
		}
	}
	p, bulkerr := client.BulkProcessor().Name("HostImporter").Workers(1).BulkActions(500).BulkSize(2 << 20).FlushInterval(30 * time.Second).Do()
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
			indexDoc := elastic.NewBulkUpdateRequest().Index("passive-ssl-sonar-hosts").Type("host").Id(id).Doc(nh).DocAsUpsert(true)
			p.Add(indexDoc)
		case <-Done:
			fmt.Println("Got done in es...flushing")
			p.Flush()
			break

		}
	}
	//bulkRequest := client.Bulk()

	elasticerr := p.Close()
	if elasticerr != nil {
		log.Println(err)
	}

}

func search_newhosts() {
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	p, bulkerr := client.BulkProcessor().Name("HostImporter").Workers(1).BulkActions(500).BulkSize(2 << 20).FlushInterval(30 * time.Second).Do()
	if bulkerr != nil {
		fmt.Println(bulkerr)
	}
	query := elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewExistsQuery("first_seen"))
	fmt.Println("Search hits are:")
	sr, err := client.Scan().Index("passive-ssl-sonar-hosts").Query(query).FetchSource(true).Do()
	//sr, err := client.Scroll().Index("passive-ssl-sonar-hosts").Query(query).Do()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sr.TotalHits())

	if sr.TotalHits() > 0 {
		fmt.Printf("Found a total of %d hosts\n", sr.TotalHits())
		for {
			res, err := sr.Next()
			if err == elastic.EOS {
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

	lookupchan := make(chan Host, 10000)
	indexchan := make(chan Host, 10000)
	var wg sync.WaitGroup
	Done := make(chan struct{})
	defer close(Done)
	fmt.Println("Starting import at: ", time.Now())
	for w := 1; w <= 3; w++ {
		go Lookup_ip(lookupchan, indexchan, Done)
	}

	wg.Add(2)

	go file_reader(lookupchan, hostsfile, wg, Done)
	go ESWriter(indexchan, wg, Done)

	wg.Wait()
	fmt.Println("Finished import at: ", time.Now())
	fmt.Println("Update first_seen started at: ", time.Now())

	// Now we need to go back and update...hopefully it works
	search_newhosts()
	fmt.Println("Update first_seen finished at: ", time.Now())

}
