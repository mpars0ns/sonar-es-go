package helpers

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

func checkCreateSonarSSLIndex()  {
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
		return
	}
	fmt.Println("The index already existed")
	return
}
