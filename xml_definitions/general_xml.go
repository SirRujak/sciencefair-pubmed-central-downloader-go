package xml_definitions

type PubmedArticleSet struct {
	PubmedArticles *[]PubmedArticle `xml:"PubmedArticleSet"`
}

type PubmedArticle struct {
	MedlineCitation MedlineCitation `xml:"MedlineCitation"`
	PubmedData      PubmedData      `xml:"PubmedData"`
}

type MedlineCitation struct {
	Status                  string                `xml:"Status,attr"`
	Owner                   string                `xml:"Owner,attr"`
	PMID                    string                `xml:"PMID,chardata"`
	PMIDVersion             string                `xml:"PMID>Version,attr"`
	DateCompleted           Date                  `xml:"DateCompleted"`
	DateRevised             Date                  `xml:"DateRevised"`
	Article                 Article               `xml:"Article"`
	MedlineJournalInfo      MedlineJournalInfo    `xml:"MedlineJournalInfo"`
	ChemicalList            []Chemical            `xml:"ChemicalList"`
	CitationSubset          string                `xml:"CitationSubset,chardata"`
	CommentsCorrectionsList []CommentsCorrections `xml:"CommentsCorrectionsList"`
	MeshHeadingList         []MeshHeading         `xml:"MeshHeadingList"`
}

type MedlineJournalInfo struct {
	Country   string `xml:"Country,chardata"`
	MedlineTA string `xml:"MedlineTA,chardata"`
	// TODO: Otherwise possibly use this for journal identifier?
	NumUniqueID int    `xml:"NumUniqueID,chardata"`
	ISSNLinking string `xml:"ISSNLinking,chardata"`
}

type Chemical struct {
	RegistryNumber  int       `xml:"RegistryNumber,chardata"`
	NameOfSubstance Substance `xml:"NameOfSubstance"`
}

type Substance struct {
	UI   string `xml:"UI,attr"`
	Name string `xml:",chardata"`
}

type CommentsCorrections struct {
	RefType   string      `xml:"RefType,attr"`
	RefSource string      `xml:"RefSource,chardata"`
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
	Year   string `xml:"Year,chardata"`
	Month  string `xml:"Month,chardata"`
	Day    string `xml:"Day,chardata"`
	Hour   string `xml:"Hour,chardata"`
	Minute string `xml:"Minute,chardata"`
}

type ArticleDate struct {
	DateType string `xml:"DateType,attr"`
	Year     string `xml:"Year,chardata"`
	Month    string `xml:"Month,chardata"`
	Day      string `xml:"Day,chardata"`
	Hour     string `xml:"Hour,chardata"`
	Minute   string `xml:"Minute,chardata"`
}

type ISSN struct {
	IssnType  string `xml:"IssnType,attr"`
	ISSNValue string `xml:",chardata"`
}

type JournalIssue struct {
	CitedMedium string `xml:"CitedMedium,attr"`
	Volume      string `xml:"Volume,chardata"`
	Issue       string `xml:"Issue,chardata"`
	PubDate     Date   `xml:"PubDate"`
}

type Journal struct {
	ISSN         ISSN         `xml:"ISSN"`
	JournalIssue JournalIssue `xml:"JournalIssue"`
	Title        string       `xml:"Title,chardata"`
	// TODO: This should be what we use to split it up by publisher.
	ISOAbbreviation string `xml:"ISOAbbreviation,chardata"`
}

type Pagination struct {
	MedlinePgns []string `xml:"MedlinePgn,chardata"`
}

type ELocationID struct {
	EIdType string `xml:"EIdType,attr"`
	ValidYN string `xml:"ValidYN,attr"`
	ID      string `xml:",chardata"`
}

type Abstract struct {
	AbstractText string `xml:"AbstractText,chardata"`
}

type AffiliationInfo struct {
	Affiliation []string `xml:"Affiliation,chardata"`
}

type Author struct {
	ValidYN         string          `xml:"ValidYN,attr"`
	LastName        string          `xml:"LastName,chardata"`
	ForeName        string          `xml:"ForeName,chardata"`
	Initials        string          `xml:"Initials,chardata"`
	AffiliationInfo AffiliationInfo `xml:"AffiliationInfo"`
}

type AuthorList struct {
	CompleteYN string   `xml:"CompleteYN,attr"`
	Authors    []Author `xml:"Author"`
}

type Grant struct {
	GrantID string `xml:"GrantID,chardata"`
	Agency  string `xml:"Agency,chardata"`
	Country string `xml:"Country,chardata"`
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
	ArticleTitle        string            `xml:"ArticleTitle,chardata"`
	Pagination          Pagination        `xml:"Pagination"`
	ELocationID         ELocationID       `xml:"ELocationID"`
	Abstract            Abstract          `xml:"Abstract"`
	AuthorList          AuthorList        `xml:"AuthorList"`
	Language            string            `xml:"Language,chardata"`
	GrantList           GrantList         `xml:"GrantList"`
	PublicationTypeList []PublicationType `xml:"PublicationType"`
	ArticleDate         ArticleDate       `xml:"ArticleDate"`
}

type PubmedData struct {
	History           []PubMedPubDate `xml:"History"`
	PublicationStatus string          `xml:"PublicationStatus,chardata"`
	ArticleIDList     []ArticleID     `xml:"ArticleIDList"`
}

type PubMedPubDate struct {
	PubStatus string `xml:"PubStatus,attr"`
	Year      string `xml:"Year,chardata"`
	Month     string `xml:"Month,chardata"`
	Day       string `xml:"Day,chardata"`
	Hour      string `xml:"Hour,chardata"`
	Minute    string `xml:"Minute,chardata"`
}

type ArticleID struct {
	IDType string `xml:"IdType,attr"`
	ID     string `xml:",chardata"`
}
