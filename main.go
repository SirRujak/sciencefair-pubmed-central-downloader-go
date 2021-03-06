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

	"./json_definitions"
	"./xml_definitions"

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
	//log.Print(url)
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
	LastDate     string `json:"last_date"`
	LastSize     int64  `json:"last_size"`
	EmailAddress string `json:"email"`
}

func readJSON(configPath string) (*config, error) {
	//const configPath = "./config.json"
	var loadedConfig config

	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Print("unable to open json file to read")
		log.Print(configPath)
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

func saveJSON(newConfig *config, configPath string) error {
	//const configPath = "./config.json"
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
		return errors.New("unable to write to file")
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

func convertXMLToJSON(xmlStruct *xml_definitions.PubmedArticle, articlePath string, doi *string, pmcid string) (*json_definitions.Metadata, error) {

	pathType := "/"
	compressionType := "tgz"

	tempJSON := json_definitions.Metadata{
		CompressionType: &compressionType,
		// This designates that it is already broken up into paths deleniated by
		// the "/" symbol.
		PathType:  &pathType,
		Path:      &articlePath,
		EntryFile: "main.nxml",
	}

	// Need to at least pull out:
	// Title
	// Abstract
	// Identifier
	// Date
	// AuthorList
	// License
	// Possibly try and grab the license with this?: It is in the nxml files so we will get it later if we need it but this should be fine for now.
	// https://www.ncbi.nlm.nih.gov/pmc/oai/oai.cgi?verb=GetRecord&identifier=oai:pubmedcentral.nih.gov:3728067&metadataPrefix=oai_dc
	tempJSON.Title = xmlStruct.MedlineCitation.Article.ArticleTitle
	tempJSON.Abstract = xmlStruct.MedlineCitation.Article.Abstract.AbstractText
	tempIdentifier := json_definitions.Identifier{
		Type: "pmid",
		ID:   xmlStruct.MedlineCitation.PMID.PMID,
	}
	tempJSON.Identifier = append(tempJSON.Identifier, tempIdentifier)
	if doi != nil {
		tempDOI := json_definitions.Identifier{
			Type: "doi",
			ID:   *doi,
		}
		tempJSON.Identifier = append(tempJSON.Identifier, tempDOI)
	}
	tempPMCID := json_definitions.Identifier{
		Type: "pmcid",
		ID:   pmcid,
	}
	tempJSON.Identifier = append(tempJSON.Identifier, tempPMCID)
	tempDate := json_definitions.Date{
		Day:   xmlStruct.MedlineCitation.DateCompleted.Day,
		Month: xmlStruct.MedlineCitation.DateCompleted.Month,
		Year:  xmlStruct.MedlineCitation.DateCompleted.Year,
	}
	tempJSON.Date = tempDate
	for author := 0; author < len(xmlStruct.MedlineCitation.Article.AuthorList.Authors); author++ {
		tempAuthor := json_definitions.Author{
			Surname:    xmlStruct.MedlineCitation.Article.AuthorList.Authors[author].LastName,
			GivenNames: xmlStruct.MedlineCitation.Article.AuthorList.Authors[author].ForeName,
		}
		tempJSON.AuthorList = append(tempJSON.AuthorList, tempAuthor)
	}

	return &tempJSON, nil
}

func downloadArticle(url string, destination string) error {
	// Download the article at url and extract it to destination.
	tempURL := url + "?archnive=false"
	os.MkdirAll(destination, 0655)
	//log.Print("Destination: " + destination)
	err := getter.Get(destination, tempURL)
	if err == nil {
		//log.Print("Downloaded " + url)
	} else {
		log.Print("Error downloading article. See error below:")
	}
	return err
}

func downloadMetaDataXML(url string) (*xml_definitions.PubmedArticleSet, error) {
	// Download the data.
	urlResponse, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer urlResponse.Body.Close()
	//log.Print(url)

	// Read the data to a []char.
	dataString, err := ioutil.ReadAll(urlResponse.Body)
	if err != nil {
		log.Print("Error reading the metadata file.")
		return nil, err
	}
	// Parse the XML data.
	pubMedMetadata := xml_definitions.PubmedArticleSet{}
	err = xml.Unmarshal(dataString, &pubMedMetadata)
	if err != nil {
		//log.Print(pubMedMetadata)
		log.Print(url)
		log.Print("issue unmarshalling xml metadata")
		log.Print(err)
		log.Print(dataString)
		return nil, err
	}
	//log.Print(pubMedMetadata)

	return &pubMedMetadata, nil
}

func downloadIDXML(url string) (*xml_definitions.PCMIDSet, error) {
	// Download the data.
	urlResponse, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer urlResponse.Body.Close()
	//log.Print(url)

	// Read the data to a []char.
	dataString, err := ioutil.ReadAll(urlResponse.Body)
	if err != nil {
		log.Print("Error reading the metadata file.")
		return nil, err
	}

	// Parse the XML data.
	pubMedMetadata := xml_definitions.PCMIDSet{}
	err = xml.Unmarshal(dataString, &pubMedMetadata)
	if err != nil {
		log.Print("issue unmarshalling id metadata")
		log.Print(err)
		return nil, err
	}
	//log.Print(pubMedMetadata)

	return &pubMedMetadata, nil
}

func downloadArticles(lastTime time.Time, updateURLBase string, articleBasePath string, metadataBasePath string, articleListing *os.File, emailAddress string, badArticleListing *os.File) error {

	var err error
	userInfo := "&tool=sciencefair_downloader&email=" + emailAddress
	const metadataBaseLink = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&retmode=XML&id="
	const pmcidBaseLink = "https://www.ncbi.nlm.nih.gov/pmc/utils/idconv/v1.0/?versions=no&idtype=pmcid&ids="
	lastTimeFormatted := lastTime.Format("2006-01-02+15:04:05")
	log.Print("lastTime=" + lastTimeFormatted)
	formatURL := "&format=tgz"
	fullUpdateURL := updateURLBase + lastTimeFormatted + formatURL
	log.Print(fullUpdateURL)
	// If there is anything in the article list download them.
	// Continue until the resumption link is nil.
	var updateComplete = false
	//getterClient := &getter.Client{}
	var numNewArticles int
	for updateComplete != true {
		var updateXML []byte
		updateXML, err = downloadUpdateXML(fullUpdateURL)
		if err != nil {
			log.Print(err)
			return err
		}

		var update databaseUpdate
		xml.Unmarshal(updateXML, &update)
		//log.Print(update)
		if update.Records.Resumption == nil {
			updateComplete = true
		} else {
			// TODO: Swap this back to fullUpdateURL when it is working
			// so it will actually download everything.
			updateComplete = true
			//fullUpdateURL = update.Records.Resumption.ResumptionLink.Href
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

		// We now need to do the following things to make sure we have the right
		// data:
		// Go through the loop once and separate the PMCID data into 200 article
		// chunks. Those will then be fed into this system:
		// https://www.ncbi.nlm.nih.gov/pmc/tools/id-converter-api/
		// Probably with the:
		// service-root?ids=PMC1193645&versions=no,
		// service-root?ids=PMC2883744&format=json
		// settings so that we get it as JSON, and so that we only get the most
		// recent version.
		// Go through the new list of PMID values and download their metadata in
		// batches. The metadata search can handle 500 artilces at a time.
		// Once this is finally complete, begin downloading and saving the
		// actual papers and once each paper is downloaded save its metadata file
		// and add the info to the listing.

		// Current limits of 3 requests per second.
		PMCIDList := []record{}
		// Create a slice of record structs.
		for PMCID := 0; PMCID < numNewArticles; PMCID++ {
			if update.Records.RecordList[PMCID].Link.Format == "pdf" {
				continue
			}

			// This means it is not a pdf entry.
			PMCIDList = append(PMCIDList, update.Records.RecordList[PMCID])
		}

		// Make a copy of the slice  and then make batches of it.
		copyPMCIDList := make([]record, len(PMCIDList))
		copy(copyPMCIDList, PMCIDList)
		var numBatches = len(PMCIDList) / 200.0

		PMCIDBatches := [][]record{}
		var singleBatch []record

		// Deal with all but the last batch.
		if numBatches > 1 {
			for i := 0; i < numBatches-1; i++ {
				singleBatch, copyPMCIDList = copyPMCIDList[:200], copyPMCIDList[200:]
				PMCIDBatches = append(PMCIDBatches, singleBatch)
			}
		}
		// Deal with the last batch.
		PMCIDBatches = append(PMCIDBatches, copyPMCIDList)
		//log.Print(PMCIDBatches)

		// Step through batches and download the ID conversion data and links.
		// If there is an issue with the ID conversion data, don't download it and
		// add it to the badPMCIDList.
		// Otherwise, add the PMCID to finalPMCIDList, the record struct to
		// finalPMCRecordList the ID conversion
		// data to fullPMIDList, the PMID to totalPMIDList, and the PMID
		// to totalPMIDList.

		totalPMIDList := make([]string, 0)
		fullPMIDList := make([]xml_definitions.Record, 0)

		finalPMCIDList := make([]string, 0)
		finalPMCRecordList := make([]record, 0)
		badPMCIDList := make([]string, 0)

		for PMCIDBatch := 0; PMCIDBatch < len(PMCIDBatches); PMCIDBatch++ {
			currentBatch := PMCIDBatches[PMCIDBatch]
			PMCIDString := []string{}
			for i := 0; i < len(currentBatch); i++ {
				PMCIDString = append(PMCIDString, currentBatch[i].ID)
			}
			metadataPCMID := strings.Join(PMCIDString[:], ",")

			// Download the PCMIDSet data.
			pmidDataURL := pmcidBaseLink + metadataPCMID + userInfo
			var articlePMIDData *xml_definitions.PCMIDSet
			articlePMIDData, err = downloadIDXML(pmidDataURL)
			//log.Print(articlePMIDData)
			if err != nil {
				return err
			}

			tempRecords := articlePMIDData.Records
			for i := 0; i < len(tempRecords); i++ {
				if tempRecords[i].PMID == "" {
					//log.Print(tempRecords[i])
					badPMCIDList = append(badPMCIDList, tempRecords[i].PMCID)
					_, err = badArticleListing.WriteString(tempRecords[i].PMCID + "," + "PMIDError" + "\n")
					if err != nil {
						log.Print("Issue getting metadata of and saving reference to: " + tempRecords[i].PMCID)
						continue
					}
				} else if tempRecords[i].DOI == "" {
					//log.Print(tempRecords[i])
					badPMCIDList = append(badPMCIDList, tempRecords[i].PMCID)
					_, err = badArticleListing.WriteString(tempRecords[i].PMCID + "," + "DOIError" + "\n")
					if err != nil {
						log.Print("Issue getting metadata of and saving reference to: " + tempRecords[i].PMCID)
						continue
					}
				} else {
					finalPMCIDList = append(finalPMCIDList, tempRecords[i].PMCID)
					finalPMCRecordList = append(finalPMCRecordList, currentBatch[i])
					fullPMIDList = append(fullPMIDList, tempRecords[i])
					totalPMIDList = append(totalPMIDList, tempRecords[i].PMID)

				}
			}
		}

		copyTotalPMIDList := make([]string, len(totalPMIDList))
		copy(copyTotalPMIDList, totalPMIDList)
		numBatches = len(totalPMIDList) / 200.0

		PMIDBatches := [][]string{}
		var singlePMIDBatch []string
		// Deal with all but the last batch.
		if numBatches > 1 {
			for i := 0; i < numBatches-1; i++ {
				singlePMIDBatch, copyTotalPMIDList = copyTotalPMIDList[:200], copyTotalPMIDList[200:]
				PMIDBatches = append(PMIDBatches, singlePMIDBatch)
			}
		}
		// Deal with the last batch.
		PMIDBatches = append(PMIDBatches, copyTotalPMIDList)

		// Step through batches and download the metadata an links.
		totalPubmedArticleList := make([]*xml_definitions.PubmedArticle, 0)

		for PMIDBatch := 0; PMIDBatch < len(PMIDBatches); PMIDBatch++ {
			// Download metadata.
			currentBatchPMIDs := PMIDBatches[PMIDBatch]
			metadataPMID := strings.Join(currentBatchPMIDs[:], ",")
			metaDataURL := metadataBaseLink + metadataPMID + userInfo
			//log.Print("test1")
			var articleMetadata *xml_definitions.PubmedArticleSet
			articleMetadata, err = downloadMetaDataXML(metaDataURL)
			if err != nil {
				return err
			}
			pubmedArticleList := *articleMetadata.PubmedArticles
			for tempArticle := 0; tempArticle < len(pubmedArticleList); tempArticle++ {
				totalPubmedArticleList = append(totalPubmedArticleList, &pubmedArticleList[tempArticle])
			}
		}

		numNewArticles = len(finalPMCIDList)

		for currentArticle := 0; currentArticle < numNewArticles; currentArticle++ {
			//log.Print(update.Records.RecordList)
			/*
				if update.Records.RecordList[currentArticle].Link.Format == "pdf" {
					continue
				}
				articleLinkFtp := update.Records.RecordList[currentArticle].Link.Href
			*/
			articleLinkFtp := finalPMCRecordList[currentArticle].Link.Href
			// Process the link to find the hashed directory names.
			articleList := strings.SplitN(articleLinkFtp, "/", 3)
			// No longer need this next line since we download everything above.
			//metadataPMID := strings.Split(strings.Split(articleLinkFtp, "PMC")[1], ".")[0]
			//log.Print(articleList)
			articleLinkHTTP := "http://" + articleList[2]
			articleListHashes := strings.Split(articleList[2], "/")
			firstHash := articleListHashes[4]
			secondHash := articleListHashes[5]

			articleDestination := []string{
				articleBasePath,
				firstHash,
				secondHash,
			}
			articlePath := path.Join(articleDestination...)
			err = downloadArticle(articleLinkHTTP, articlePath)
			if err != nil {
				log.Fatal(err)
				return err
			}

			// Download and save the article metadata using:
			// https://www.ncbi.nlm.nih.gov/pmc/tools/get-metadata/
			// Specifically:
			// https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi?db=pubmed&id=PMID
			// Where PMID is the PMID of the article.

			hashPath := path.Join(firstHash, secondHash)
			/*
				// No longer use this because we downloaded this information before.
				metaDataURL := metadataBaseLink + metadataPMID + userInfo
				articleMetadata, err := downloadMetaDataXML(metaDataURL)
				if err != nil {
					return err
				}

				// Convert the downloaded data to the sciencefair JSON format.
				// Start by filling in the defaults for this repository.
				//log.Print(articleMetadata)
				if articleMetadata == nil {

				}
				singleArticle := *articleMetadata.PubmedArticles
			*/
			singleArticle := totalPubmedArticleList[currentArticle]
			metadataJSON, err := convertXMLToJSON(singleArticle, hashPath, &fullPMIDList[currentArticle].DOI, finalPMCIDList[currentArticle])
			if err != nil {
				log.Print("issue converting xml to json")
				return err
			}
			metadataString, err := json.Marshal(metadataJSON)
			if err != nil {
				log.Print("issue marshalling to json")
			}

			// Save the metadata string to a json file.
			// Use name PubMedCentral-PMID-v2.json
			metadataDestination := []string{
				metadataBasePath,
				firstHash,
				secondHash,
			}

			// Create the path to the json file.
			metadataPath := path.Join(metadataDestination...)

			err = os.MkdirAll(metadataPath, 0655)
			if err != nil {
				log.Print("issue creating directories for metadata files")
				return err
			}

			metadataFileName := path.Join(metadataPath, "PubMedCentral-"+metadataJSON.Identifier[0].ID+"-v2.json")
			err = ioutil.WriteFile(metadataFileName, metadataString, 0655)
			if err != nil {
				log.Print("issue saving metadata json file")
				return err
			}

			// Save the following to articleListing:
			// key, value1, value2 sets on each line of
			// KEY = ARTICLE_IDENTIFIER (PMID)
			// VALUE1 = PATH_TO_ARTICLE (Same path as articlePath)
			// VALUE2 = TIME_OF_ARTICLE_UPDATE (Found in the article metadata.)
			// VALUE3 = PMCID
			// VALUE4 = DOI
			csvData := metadataJSON.Identifier[0].ID + "," + firstHash +
				"/" + secondHash + "," + metadataJSON.Date.Year +
				metadataJSON.Date.Month + metadataJSON.Date.Day + "," +
				fullPMIDList[currentArticle].PMCID + "," +
				fullPMIDList[currentArticle].DOI +
				"\n"
			_, err = articleListing.WriteString(csvData)
			if err != nil {
				log.Print("issue writing to csv index")
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
	pwd = path.Join(pwd, "PMCData")
	articleBasePath := path.Join(pwd, "articles")
	metadataBasePath := path.Join(pwd, "metadata")
	oafilesPath := path.Join(pwd, "oa_files")
	configPath := path.Join(pwd, "config.json")
	articleListingPath := path.Join(oafilesPath, "article_listing.csv")
	badArticleListingPath := path.Join(oafilesPath, "bad_article_listing.csv")
	//log.Print(articleListingPath)
	//var firstRun bool
	//firstRun = false
	var err error
	lastConfig, err := readJSON(configPath)
	if err != nil {
		log.Print("uanble to load json file")
		lastConfig = &config{}
	}
	var files []os.FileInfo
	files, err = ioutil.ReadDir(oafilesPath)
	if err != nil {
		log.Print("Unable to read oa_files. Creating folder.", err)
		os.MkdirAll(oafilesPath, 0655)
		files, err = ioutil.ReadDir(oafilesPath)
		if err != nil {
			log.Print("Still unable to create folder.")
			panic(err)
		}
	}

	// Check if any files were found.
	//const initialURL = "http://ftp.ncbi.nlm.nih.gov/pub/pmc/oa_file_list.csv"
	const updateURLBase = "https://www.ncbi.nlm.nih.gov/pmc/utils/oa/oa.fcgi?from="
	currentTime := time.Now()
	lastTime, err := time.Parse("20060102150405", lastConfig.LastDate)
	if err != nil {
		log.Print("Unable to load last time. Assuming not dealing with updates.")
		//lastTime = currentTime
		lastTime, err = time.Parse("20060102150405", "20000101000000")
		if err != nil {
			log.Print("Issue parsing time.")
			panic(err)
		}
	}
	emailAddress := lastConfig.EmailAddress

	// Open a csv file to place key, value1, value2 sets on each line of
	// KEY = ARTICLE_IDENTIFIER
	// VALUE1 = PATH_TO_ARTICLE
	// VALUE2 = TIME_OF_ARTICLE_UPDATE
	//if len(files) <= 0 || currentTime.Unix() > lastTime.Add(24*time.Hour).Unix() {
	if len(files) <= 0 {
		// Download new file.
		//os.MkdirAll(path, perm)
		//log.Print(articleListingPath)
		articleListing, err := os.OpenFile(articleListingPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0655)
		if err != nil {
			log.Print("Issue opening or creating article listing file. Permission error?")
			panic(err)
		}
		defer articleListing.Close()

		badArticleListing, err := os.OpenFile(badArticleListingPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0655)
		if err != nil {
			log.Print("Issue opening or creating article listing file. Permission error?")
			panic(err)
		}
		defer badArticleListing.Close()

		log.Print("Downloading because we do not yet have a file.")
		/*
			err = downloadXML(initialURL)
			if err != nil {
				log.Print("Issue downloading the XML.", err)
				return
			}
		*/
		err = downloadArticles(lastTime, updateURLBase, articleBasePath, metadataBasePath, articleListing, emailAddress, badArticleListing)
		if err != nil {
			panic(err)
		} else {
			log.Print("No errors!")
			return
		}
	} else if currentTime.Unix() > lastTime.Add(24*time.Hour).Unix() {
		// NEW STUFF
		tempPath := path.Join([]string{pwd, "oa_files", "article_listing.csv"}...)
		articleListing, err := os.OpenFile(tempPath, os.O_RDWR|os.O_APPEND, 0655)
		if err != nil {
			log.Print("Issue opening article listing file. Permission error?")
			panic(err)
		}
		defer articleListing.Close()

		badArticleListing, err := os.OpenFile(badArticleListingPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0655)
		if err != nil {
			log.Print("Issue opening or creating article listing file. Permission error?")
			panic(err)
		}
		defer badArticleListing.Close()

		log.Print("Downloading because it has been more than 24 hours since last update.")
		err = downloadArticles(lastTime, updateURLBase, articleBasePath, metadataBasePath, articleListing, emailAddress, badArticleListing)
		if err != nil {
			panic(err)
		} else {
			// TODO: Save the time this was started into the json file so that we
			// can continue.
			log.Print("Update complete!")
		}
	} else {
		log.Print("No changes detected. Exiting...")
		return
	}
}
