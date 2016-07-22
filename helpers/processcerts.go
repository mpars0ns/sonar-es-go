package helpers

import ()
import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/pem"
	"fmt"
	"github.com/zmap/zgrab/ztools/x509"
	"gopkg.in/olivere/elastic.v3"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

func certsFileReader(indexchan *chan []string, certsfile string, Done chan struct{}) {

	fmt.Println(certsfile)
	f, err := os.Open(certsfile)
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
		select {
		case *indexchan <- data:
		case <-Done:
			return
		}
	}
}

func CertsESWriter(indexchan *chan []string, esWg *sync.WaitGroup, Done chan struct{}) {
	defer esWg.Done()

	startcert := "-----BEGIN CERTIFICATE-----\n"
	endcert := "\n-----END CERTIFICATE-----"

	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal("error connecting to ES", err)
	}

	p, bulkerr := client.BulkProcessor().Name("CertsImporter").Workers(1).BulkActions(1000).BulkSize(2 << 20).FlushInterval(30 * time.Second).Do()
	if bulkerr != nil {
		fmt.Println("Problem with the elastic bulk importer", bulkerr)
	}
OuterLoop:
	for {
		select {
		case data := <-*indexchan:
			sha1 := data[0]
			cert := startcert + data[1] + endcert
			var certblock *pem.Block
			certblock, _ = pem.Decode([]byte(cert))
			newcert, err := x509.ParseCertificate(certblock.Bytes)
			if err != nil {
				//log.Println(sha1, err)
				err = p.Flush()
				if err != nil {
					log.Println(err)
				}
				continue
			}
			//pc, _ := json.Marshal(newcert)
			indexDoc := elastic.NewBulkUpdateRequest().Index("passive-ssl-sonar-certs").Type("certificate").Id(sha1).Doc(newcert).DocAsUpsert(true)
			//fmt.Println(indexDoc)
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

func Process_Certs(certsfile string) {
	checkCreateSonarCertsSSLIndex()
	indexchan := make(chan []string, 20000)
	Done := make(chan struct{})
	var esWg sync.WaitGroup

	esWg.Add(1)
	go CertsESWriter(&indexchan, &esWg, Done)
	fmt.Println("Starting import at: ", time.Now())
	certsFileReader(&indexchan, certsfile, Done)
	close(Done)
	esWg.Wait()
	fmt.Println("Finished import at: ", time.Now())
}
