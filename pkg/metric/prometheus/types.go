package prometheus

type Annotations struct {
	Scrape bool
	Scheme string
	Path   string
	Port   int
}
