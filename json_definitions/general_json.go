package json_definitions

type Metadata struct {
	Title           string     `json:"title"`
	AuthorList      []Author   `json:"author"`
	Abstract        string     `json:"abstract"`
	Identifier      Identifier `json:"identifier"`
	Date            Date       `json:"date"`
	License         *string    `json:"license"`
	Path            *string    `json:"path"`
	EntryFile       string     `json:"entryfile"`
	Files           *[]string  `json:"files"`
	PathType        *string    `json:"path-type"`
	CompressionType *string    `json:"compression-type"`
}

type Author struct {
	Surname    string `json:"surname"`
	GivenNames string `json:"given-names"`
}

type Identifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Date struct {
	Day   string `json:"day"`
	Month string `json:"month"`
	Year  string `json:"year"`
}
