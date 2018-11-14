package xml_definitions

import "encoding/xml"

type PCMIDSet struct {
	XMLName xml.Name `xml:"pmcids"`
	Status  string   `xml:"status,attr"`
	Records []Record `xml:"record"`
}

type Record struct {
	RequestID string `xml:"requested-id,attr"`
	PMCID     string `xml:"pmcid,attr"`
	PMID      string `xml:"pmid,attr"`
	DOI       string `xml:"doi,attr"`
}
