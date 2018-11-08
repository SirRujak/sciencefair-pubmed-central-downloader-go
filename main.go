package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-getter"
)

func downloadXML(url string) error {
	// Build the file path.
	pwd, _ := os.Getwd()
	basePath := path.Join(pwd, "oa_files", "oa_file_list_")
	const baseEnd = ".csv"
	//const url = "http://ftp.ncbi.nlm.nih.gov/pub/pmc/oa_file_list.csv"

	filePath := basePath + time.Now().Format("20060102150405") + baseEnd

	// Create a file to put the data in.
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Download the data.
	urlResponse, err := http.Get(url)
	if err != nil {
		os.Remove(filePath)
		return err
	}

	defer urlResponse.Body.Close()

	// Write the data to the file.
	_, err = io.Copy(outFile, urlResponse.Body)
	if err != nil {
		os.Remove(filePath)
		return err
	}

	return nil
}

func downloadUpdateXML(url string) ([]byte, error) {
	log.Print(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Status error on: " + url + " Code: " + string(resp.StatusCode))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Read Body error on :" + url)
	}

	return data, nil
}

type config struct {
	LastDate string `json:"last_date"`
	LastSize int64  `json:"last_size"`
}

func readJSON() (*config, error) {
	const configPath = "./config.json"
	var loadedConfig config

	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Print("unable to open json file to read")
		return nil, errors.New("unable to open json file to read")
	}
	defer jsonFile.Close()

	byteArray, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.New("unable to read json file")
	}

	json.Unmarshal(byteArray, &loadedConfig)

	return &loadedConfig, nil
}

func saveJSON(newConfig *config) error {
	const configPath = "./config.json"
	jsonString, err := json.Marshal(newConfig)
	if err != nil {
		log.Print("Unable to marshal config data.")
		return err
	}

	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Print("unable to open json file to write")
		return errors.New("unable to open json file to write")
	}
	defer jsonFile.Close()

	err = ioutil.WriteFile(configPath, jsonString, 0644)
	if err != nil {
		return errors.New("Unable to write to file.")
	}
	return nil
}

type article struct {
	// Use PMID for identifying unique articles.
	File            string `json:"File"`
	ArticleCitation string `json:"ArticleCitation"`
	AccessionID     string `json:"AccessionID"`
	LastUpdated     string `json:"LastUpdated"`
	PMID            string `json:"PMID"`
	License         string `json:"License"`
}

type request struct {
	From string `xml:"from,attr"`
	Data string `xml:",chardata"`
}

type resumptionLink struct {
	Token string `xml:"token,attr"`
	Href  string `xml:"href,attr"`
}

type resumption struct {
	ResumptionLink resumptionLink `xml:"link"`
}

type recordLink struct {
	Format  string `xml:"format,attr"`
	Updated string `xml:"updated,attr"`
	Href    string `xml:"href,attr"`
}

type record struct {
	ID       string     `xml:"id,attr"`
	Citation string     `xml:"citation,attr"`
	Link     recordLink `xml:"link"`
}

type records struct {
	ReturnedCount string      `xml:"returned-count,attr"`
	TotalCount    string      `xml:"total-count,attr"`
	Resumption    *resumption `xml:"resumption"`
	RecordList    []record    `xml:"record"`
}

type databaseUpdate struct {
	// Used to store XML data retrieved from here:
	// https://www.ncbi.nlm.nih.gov/pmc/tools/oa-service/
	ResponseDate string  `xml:"responseDate"`
	Request      request `xml:"request"`
	Records      records `xml:"records"`
}

func downloadArticle(url string, destination string) error {
	// Download the article at url and extract it to destination.
	tempURL := url + "?archinve=false"
	os.MkdirAll(destination, 0655)
	log.Print("Destination: " + destination)
	err := getter.Get(destination, tempURL)
	if err == nil {
		log.Print("Downloaded " + url)
	} else {
		log.Print("Error downloading article. See error below:")
	}
	return err
}

func downloadArticles(lastTime time.Time, updateURLBase string, articleBasePath string) error {
	var err error
	log.Print("Downloading because it has been more than 24 hours since last update.")
	lastTimeFormatted := lastTime.Format("2006-01-02+15:04:05")
	formatURL := "&format=tgz"
	fullUpdateURL := updateURLBase + lastTimeFormatted + formatURL
	log.Print("test2")
	// TODO: Retrieve the XML update.
	// If there is anything in the article list download them.
	// Continue until the resumption link is nil.
	var updateComplete bool = false
	//getterClient := &getter.Client{}
	var numNewArticles int
	for updateComplete != true {
		var updateXML []byte
		log.Print("test")
		updateXML, err = downloadUpdateXML(fullUpdateURL)
		if err != nil {
			log.Print(err)
			return err
		}

		var update databaseUpdate
		xml.Unmarshal(updateXML, &update)
		log.Print("test3")
		log.Print(update)
		if update.Records.Resumption == nil {
			updateComplete = true
		} else {
			fullUpdateURL = update.Records.Resumption.ResumptionLink.Href
		}
		numNewArticles, err = strconv.Atoi(update.Records.ReturnedCount)
		if err != nil {
			log.Print("Issue discovering the number of articles to download. Terminating...")
			return err
		}
		if numNewArticles <= 0 {
			log.Print("Somehow there were no new articles in this file. Terminating...")
			return err
		}
		log.Print("test4")
		for currentArticle := 0; currentArticle < numNewArticles; currentArticle++ {
			// TODO: Download the article and save it.
			//log.Print(update.Records.RecordList)
			if update.Records.RecordList[currentArticle].Link.Format == "pdf" {
				continue
			}
			articleLinkFtp := update.Records.RecordList[currentArticle].Link.Href
			// Process the link to find the hashed directory names.
			articleList := strings.SplitN(articleLinkFtp, "/", 3)
			log.Print(articleList)
			articleLinkHttp := "http://" + articleList[2]
			//log.Print("test6")
			articleListHashes := strings.Split(articleList[2], "/")
			firstHash := articleListHashes[4]
			secondHash := articleListHashes[5]
			//log.Print("test5")

			articleDestination := []string{
				articleBasePath,
				firstHash,
				secondHash,
			}
			articlePath := path.Join(articleDestination...)
			err = downloadArticle(articleLinkHttp, articlePath)
			if err != nil {
				log.Fatal(err)
				return err
			}
		}
	}
	log.Print("Update complete!")
	return nil
}

func main() {
	// Read the oa_files folder to see if there is a previously downloaded
	// listing.
	pwd, _ := os.Getwd()
	articleBasePath := path.Join(pwd, "articles")
	//var firstRun bool
	//firstRun = false
	var err error
	lastConfig, err := readJSON()
	if err != nil {
		log.Print("uanble to load json file")
		lastConfig = &config{}
	}
	var files []os.FileInfo
	files, err = ioutil.ReadDir("./oa_files")
	if err != nil {
		log.Fatal("Unable to read oa_files. 1.", err)
	}

	// Check if any files were found.
	const initialURL = "http://ftp.ncbi.nlm.nih.gov/pub/pmc/oa_file_list.csv"
	const updateURLBase = "https://www.ncbi.nlm.nih.gov/pmc/utils/oa/oa.fcgi?from="
	currentTime := time.Now()
	lastTime, err := time.Parse("20060102150405", lastConfig.LastDate)
	if err != nil {
		log.Print("Unable to load last time. Assuming not dealing with updates.")
		lastTime = currentTime
	}

	// TODO: Open a csv file to place key, value1, value2 sets on each line of
	// KEY = ARTICLE_IDENTIFIER
	// VALUE1 = PATH_TO_ARTICLE
	// VALUE2 = TIME_OF_ARTICLE_UPDATE
	//if len(files) <= 0 || currentTime.Unix() > lastTime.Add(24*time.Hour).Unix() {
	if len(files) <= 0 {
		// Download new file.
		log.Print("Downloading because we do not yet have a file.")
		err = downloadXML(initialURL)
		if err != nil {
			log.Print("Issue downloading the XML.", err)
			return
		}
	} else if currentTime.Unix() > lastTime.Add(24*time.Hour).Unix() {
		log.Print("Downloading because it has been more than 24 hours since last update.")
		lastTimeFormatted := lastTime.Format("2006-01-02+15:04:05")
		formatURL := "&format=tgz"
		fullUpdateURL := updateURLBase + lastTimeFormatted + formatURL
		log.Print("test2")
		// If there is anything in the article list download them.
		// Continue until the resumption link is nil.
		var updateComplete bool = false
		//getterClient := &getter.Client{}
		var numNewArticles int
		for updateComplete != true {
			var updateXML []byte
			log.Print("test")
			updateXML, err = downloadUpdateXML(fullUpdateURL)
			if err != nil {
				log.Print(err)
				return
			}

			var update databaseUpdate
			xml.Unmarshal(updateXML, &update)
			log.Print("test3")
			log.Print(update)
			if update.Records.Resumption == nil {
				updateComplete = true
			} else {
				fullUpdateURL = update.Records.Resumption.ResumptionLink.Href
			}
			numNewArticles, err = strconv.Atoi(update.Records.ReturnedCount)
			if err != nil {
				log.Print("Issue discovering the number of articles to download. Terminating...")
				return
			}
			if numNewArticles <= 0 {
				log.Print("Somehow there were no new articles in this file. Terminating...")
				return
			}
			log.Print("test4")
			for currentArticle := 0; currentArticle < numNewArticles; currentArticle++ {
				//log.Print(update.Records.RecordList)
				if update.Records.RecordList[currentArticle].Link.Format == "pdf" {
					continue
				}
				articleLinkFtp := update.Records.RecordList[currentArticle].Link.Href
				// Process the link to find the hashed directory names.
				articleList := strings.SplitN(articleLinkFtp, "/", 3)
				log.Print(articleList)
				articleLinkHttp := "http://" + articleList[2]
				//log.Print("test6")
				articleListHashes := strings.Split(articleList[2], "/")
				firstHash := articleListHashes[4]
				secondHash := articleListHashes[5]
				//log.Print("test5")

				articleDestination := []string{
					articleBasePath,
					firstHash,
					secondHash,
				}
				articlePath := path.Join(articleDestination...)
				err = downloadArticle(articleLinkHttp, articlePath)
				if err != nil {
					log.Fatal(err)
					return
				}
			}
		}
		log.Print("Update complete!")
		return
	} else {
		log.Print("No changes detected. Exiting...")
		return
	}
}
