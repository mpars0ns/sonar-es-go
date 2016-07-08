package helpers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type SonarStudies struct {
	Studies []struct {
		Status    string   `json:"status"`
		LongDesc  string   `json:"long_desc"`
		Name      string   `json:"name"`
		ShortDesc string   `json:"short_desc"`
		UpdatedAt string   `json:"updated_at,omitempty"`
		Authors   []string `json:"authors"`
		CreatedAt string   `json:"created_at,omitempty"`
		Uniqid    string   `json:"uniqid"`
		Tags      []string `json:"tags,omitempty"`

		Files []struct {
			Size        string `json:"size"`
			UpdatedAt   string `json:"updated-at"`
			Description string `json:"description"`
			Name        string `json:"name"`
			Fingerprint string `json:"fingerprint"`
		} `json:"files"`

		Study struct {
			URL     string `json:"url"`
			Venue   string `json:"venue"`
			Name    string `json:"name"`
			Bibtext string `json:"bibtext"`
		} `json:"study,omitempty"`

		Contact struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"contact"`

		Organization struct {
			Website string `json:"website"`
			Name    string `json:"name"`
		} `json:"organization"`
	} `json:"studies"`
}

func DownloadFeed() (s SonarStudies) {
	studies := &SonarStudies{}
	resp, err := http.Get("https://scans.io/json")
	if err != nil {
		log.Fatal("Had an error accessing sonar: ", resp.Status)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	httperr := json.Unmarshal(body, studies)
	if httperr != nil {
		log.Fatal(httperr)
	}
	return *studies
}

func DownloadFile(url string, filename string) (err error) {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	log.Printf("Started downloading %v \n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Had an error downloading file from sonar url: ", url, resp.Status)
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	log.Printf("Finished downloading %v \n", url)
	return nil
}
