package sonar_helpers

import (
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"log"
)

func Check_index_and_create(index string) bool {
	mapping := `{"settings":{"number_of_shards":1,"number_of_replicas":0}}`
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
		return false
	}
	fmt.Printf("Checking for index: %v now \n", index)
	exists, err := client.IndexExists(index).Do()
	if err != nil {
		log.Fatal(err)
		return false
	}
	if !exists {
		_, err = client.CreateIndex(index).BodyString(mapping).Do()
		if err != nil {
			log.Fatal(err)
			return false
		}
	}
	return true
}

/*
func ProccesHosts(indexchan chan *Host ) {
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal("Error connecting in ProcessHosts from elastic ", err)
	}
	p, bulkerr := client.BulkProcessor().Name("HostImporter").Workers(1).BulkActions(500).BulkSize(2 << 20).FlushInterval(30 * time.Second).Do()
	fmt.Println(bulkerr)
	if bulkerr != nil {
		fmt.Println(bulkerr)
	}
	for {
		newhost := <-indexchan
		hasher := sha1.New()
		hash_string := newhost.Host + newhost.Hash + newhost.Source
		hasher.Write([]byte(hash_string))
		id := hex.EncodeToString(hasher.Sum(nil))
		nh, _ := json.Marshal(newhost)
		fmt.Printf("Uploading host with id %v, %v", id, string(nh))
		indexDoc := elastic.NewBulkUpdateRequest().Index("passive-ssl-sonar-hosts").Type("host").Id(id).Doc(newhost).DocAsUpsert(true)
		p.Add(indexDoc)
	}

	elasticerr := p.Close()
	if elasticerr != nil {
		log.Println(err)
	}
}*/
