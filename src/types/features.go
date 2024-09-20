package types

type Match struct {
	Id             string
	FeatureName    string
	MatchType      string
	Type           string
	FoundId        bool
	MatchContent   string
	FeatureContent string
	DefaultContent string
	DelimeterStart string
	DelimeterEnd   string
}

type Delimeter struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Delimeters map[string]Delimeter

type Feature struct {
	Name  string
	State string
}

type BlockFeature struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Synced      bool   `json:"synced"`
	SwapContent string `json:"swapContent"`
}

type CommitFeature struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type Conflict struct {
	Resolved  bool
	LineStart int
	LineEnd   int
	Content   string
}

type FilePathCategory struct {
	Path   string
	Action []string
}