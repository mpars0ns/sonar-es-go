package helpers

import (
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"log"
	"time"
)

func Check_index_and_create(index string) bool {
	mapping := `{"settings":{"number_of_shards":1,"number_of_replicas":0}}`
	client, err := elastic.NewClient()
	if err != nil {
		log.Println("error connecting to ES", err)
		return false
	}
	fmt.Printf("Checking for index: %v now \n", index)
	exists, err := client.IndexExists(index).Do()
	if err != nil {
		log.Println("error checking if index exists", err)
		return false
	}
	if !exists {
		_, err = client.CreateIndex(index).BodyString(mapping).Do()
		if err != nil {
			log.Println("erorr creating index", err)
			return false
		}
		fmt.Println("Sleeping to allow ES to allocate indexes")
		time.Sleep(5 * time.Second)
	}

	return true
}

func checkCreateSonaHostsSSLIndex() {
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
		fmt.Println("Index Created")
		fmt.Println("Sleeping to allow ES to allocate indexes")
		time.Sleep(5 * time.Second)
		return
	}
	fmt.Println("The index already existed")
	return
}



func checkCreateSonarCertsSSLIndex() {
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	//let's check if index exists:
	exists, err := client.IndexExists("passive-ssl-sonar-certs").Do()
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		mapping := `{
    "settings":{
        "number_of_shards":5,
        "number_of_replicas":0
    }
}`
		_, err = client.CreateIndex("passive-ssl-sonar-certs").BodyString(mapping).Do()
		if err != nil {
			panic(err)
		}
		fmt.Println("Index Created")
		fmt.Println("Sleeping to allow ES to allocate indexes")
		time.Sleep(5 * time.Second)
		return
	}
	fmt.Println("The index already existed")
	return
}