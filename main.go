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

	/*

		// If there are two files make sure the second one is the most recent.
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})

		var sliceLength int = len(files)

		// If there is more than one file check to see if there is any new data.
		// If there is only one item in the slice we have not yet downloaded anything.
		if sliceLength > 1 {
			var lengthDifference int64 = files[sliceLength-1].Size() - files[sliceLength-2].Size()

			if lengthDifference < 0 {
				log.Print("negative difference found. need to check if some were deleted.")
				return
			} else if lengthDifference == 0 {
				log.Print("no change found.")
				//return
			}
		}

		// We need to download the files and place them in a directory tree.
		files, err = ioutil.ReadDir("./oa_files")
		log.Print("test", files)
		if err != nil {
			log.Print("Issue reading oa_files.", err)
			return
		}

		// If there are two files make sure the second one is the most recent.
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})

		// If there is only one file read all of it. Otherwise, after the file has
		// be loaded, seek to the previous EOF location.
		var inFilePath string
		inFilePath = path.Join("./oa_files", files[len(files)-1].Name())
		if len(files) == 1 {
			firstRun = true
		}

		csvFile, err := os.Open(inFilePath)
		if err != nil {
			log.Print("Issue loading csvFile.", err)
		}
		defer csvFile.Close()
		log.Print("test2")

		if firstRun != true {
			oldEOF := files[len(files)-2].Size()
			_, err = csvFile.Seek(oldEOF, 1)
			if err != nil {
				log.Fatal("Unable to seek in csv file.", err)

			}
		}

		// Update the config.json file to ensure it is the most recent.
		inFileSize := files[len(files)-1].Size()
		newConfig := config{
			LastDate: currentTime.Format("20060102150405"),
			LastSize: inFileSize,
		}
		err = saveJSON(&newConfig)

		csvReader := csv.NewReader(bufio.NewReader(csvFile))

		// TODO: Delete all of the old csv files.

		// The base URL is actually:
		// ftp://ftp.ncbi.nlm.nih.gov/pub/pmc/oa_package/
		// but since the name of each article in the csv has oa_package it is left
		// off of the end.
		const articleBaseURL = "ftp://ftp.ncbi.nlm.nih.gov/pub/pmc/"

		var tempArticle article
		var numArticles int64
		numArticles = 0

		for {
			record, err := csvReader.Read()
			// Stop for EOF.
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("Error reading line from csvReader.", err)
			}

			// Use the current line to download and save the given article.
			//fmt.Println(record)
			tempArticle.File = record[0]
			tempArticle.ArticleCitation = record[1]
			tempArticle.AccessionID = record[2]
			tempArticle.LastUpdated = record[3]
			tempArticle.PMID = record[4]
			tempArticle.License = record[5]

			numArticles += 1

			// Split the PMID into a list of strings.
			filePathArticleString := strings.Split(tempArticle.File, ".")[0]
			filePathArticleStrings := strings.Split(filePathArticleString, "/")
			pmidPathArticleStrings := filePathArticleStrings[1:len(filePathArticleStrings)]

			//pmidPathArticleStrings := strings.Split(tempArticle.PMID, "")
			pmidPathFullStrings := append([]string{articleBasePath}, pmidPathArticleStrings...)

			pmidPath := path.Join(pmidPathFullStrings...)

			os.MkdirAll(pmidPath, 0655)
			if numArticles > 10 {
				break
			}

		}

		//for _, fileName := range files {
		//	fmt.Println(fileName.Name())
		//	fmt.Println(fileName.Size())
		//}

		// TODO: Need to update the config.json with the new file information and
		// delete everything but the last file to save space.

		log.Print("Done Processing Files. Number of articles: ", numArticles)
	*/
}
