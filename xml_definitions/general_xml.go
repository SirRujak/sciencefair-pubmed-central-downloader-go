package xml_definitions

import "encoding/xml"

type PubmedArticleSet struct {
	XMLName        xml.Name         `xml:"PubmedArticleSet"`
	PubmedArticles *[]PubmedArticle `xml:"PubmedArticle"`
}

type PubmedArticle struct {
	MedlineCitation MedlineCitation `xml:"MedlineCitation"`
	PubmedData      PubmedData      `xml:"PubmedData"`
}

type MedlineCitation struct {
	Status                  string                `xml:"Status,attr"`
	Owner                   string                `xml:"Owner,attr"`
	PMID                    PMID                  `xml:"PMID"`
	DateCompleted           Date                  `xml:"DateCompleted"`
	DateRevised             Date                  `xml:"DateRevised"`
	Article                 Article               `xml:"Article"`
	MedlineJournalInfo      MedlineJournalInfo    `xml:"MedlineJournalInfo"`
	ChemicalList            []Chemical            `xml:"ChemicalList"`
	CitationSubset          string                `xml:"CitationSubset"`
	CommentsCorrectionsList []CommentsCorrections `xml:"CommentsCorrectionsList"`
	MeshHeadingList         []MeshHeading         `xml:"MeshHeadingList"`
}

type PMID struct {
	PMID        string `xml:",chardata"`
	PMIDVersion string `xml:"Version,attr"`
}

type MedlineJournalInfo struct {
	Country   string `xml:"Country"`
	MedlineTA string `xml:"MedlineTA"`
	// TODO: Otherwise possibly use this for journal identifier?
	NumUniqueID int    `xml:"NumUniqueID"`
	ISSNLinking string `xml:"ISSNLinking"`
}

type Chemical struct {
	RegistryNumber  int       `xml:"RegistryNumber"`
	NameOfSubstance Substance `xml:"NameOfSubstance"`
}

type Substance struct {
	UI   string `xml:"UI,attr"`
	Name string `xml:",chardata"`
}

type CommentsCorrections struct {
	RefType   string      `xml:"RefType,attr"`
	RefSource string      `xml:"RefSource"`
	PMID      CommentPMID `xml:"PMID"`
}

type CommentPMID struct {
	Version string `xml:"Version,attr"`
	PMID    string `xml:",chardata"`
}

type MeshHeading struct {
	DescriptorName DescriptorName `xml:"DescriptorName"`
	QualifierName  QualifierName  `xml:"QualifierName"`
}

type DescriptorName struct {
	UI           string `xml:"UI,attr"`
	MajorTopicYN string `xml:"MajorTopicYN,attr"`
	Name         string `xml:",chardata"`
}

type QualifierName struct {
	UI           string `xml:"UI,attr"`
	MajorTopicYN string `xml:"MajorTopicYN,attr"`
	Name         string `xml:",chardata"`
}

type Date struct {
	Year   string `xml:"Year"`
	Month  string `xml:"Month"`
	Day    string `xml:"Day"`
	Hour   string `xml:"Hour"`
	Minute string `xml:"Minute"`
}

type ArticleDate struct {
	DateType string `xml:"DateType,attr"`
	Year     string `xml:"Year"`
	Month    string `xml:"Month"`
	Day      string `xml:"Day"`
	Hour     string `xml:"Hour"`
	Minute   string `xml:"Minute"`
}

type ISSN struct {
	IssnType  string `xml:"IssnType,attr"`
	ISSNValue string `xml:",chardata"`
}

type JournalIssue struct {
	CitedMedium string `xml:"CitedMedium,attr"`
	Volume      string `xml:"Volume"`
	Issue       string `xml:"Issue"`
	PubDate     Date   `xml:"PubDate"`
}

type Journal struct {
	ISSN         ISSN         `xml:"ISSN"`
	JournalIssue JournalIssue `xml:"JournalIssue"`
	Title        string       `xml:"Title"`
	// TODO: This should be what we use to split it up by publisher.
	ISOAbbreviation string `xml:"ISOAbbreviation"`
}

type Pagination struct {
	MedlinePgns []string `xml:"MedlinePgn"`
}

type ELocationID struct {
	EIdType string `xml:"EIdType,attr"`
	ValidYN string `xml:"ValidYN,attr"`
	ID      string `xml:",chardata"`
}

type Abstract struct {
	AbstractText string `xml:"AbstractText"`
}

type AffiliationInfo struct {
	Affiliation []string `xml:"Affiliation"`
}

type Author struct {
	ValidYN         string          `xml:"ValidYN,attr"`
	LastName        string          `xml:"LastName"`
	ForeName        string          `xml:"ForeName"`
	Initials        string          `xml:"Initials"`
	AffiliationInfo AffiliationInfo `xml:"AffiliationInfo"`
}

type AuthorList struct {
	CompleteYN string   `xml:"CompleteYN,attr"`
	Authors    []Author `xml:"Author"`
}

type Grant struct {
	GrantID string `xml:"GrantID"`
	Agency  string `xml:"Agency"`
	Country string `xml:"Country"`
}

type GrantList struct {
	CompleteYN string  `xml:"CompleteYN,attr"`
	Grants     []Grant `xml:"Grant"`
}

type PublicationType struct {
	UI   string `xml:"UI,attr"`
	Type string `xml:",chardata"`
}

type Article struct {
	PubModel            string            `xml:"PubModel,attr"`
	Journal             Journal           `xml:"Journal"`
	ArticleTitle        string            `xml:"ArticleTitle"`
	Pagination          Pagination        `xml:"Pagination"`
	ELocationID         ELocationID       `xml:"ELocationID"`
	Abstract            Abstract          `xml:"Abstract"`
	AuthorList          AuthorList        `xml:"AuthorList"`
	Language            string            `xml:"Language"`
	GrantList           GrantList         `xml:"GrantList"`
	PublicationTypeList []PublicationType `xml:"PublicationType"`
	ArticleDate         ArticleDate       `xml:"ArticleDate"`
}

type PubmedData struct {
	History           []PubMedPubDate `xml:"History"`
	PublicationStatus string          `xml:"PublicationStatus"`
	ArticleIDList     []ArticleID     `xml:"ArticleIDList"`
}

type PubMedPubDate struct {
	PubStatus string `xml:"PubStatus,attr"`
	Year      string `xml:"Year"`
	Month     string `xml:"Month"`
	Day       string `xml:"Day"`
	Hour      string `xml:"Hour"`
	Minute    string `xml:"Minute"`
}

type ArticleID struct {
	IDType string `xml:"IdType,attr"`
	ID     string `xml:",chardata"`
}
