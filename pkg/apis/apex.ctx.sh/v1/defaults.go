package v1

// defaulted sets the scraper resource defaults
func defaulted(scraper *Scraper) {
	defaultedSpec(&scraper.Spec)
}

func defaultedSpec(spec *ScraperSpec) {
	if spec.AnnotationPrefix == nil {
		spec.AnnotationPrefix = new(string)
		*spec.AnnotationPrefix = "prometheus.io"
	}

	if spec.ScrapeIntervalSeconds == nil {
		spec.ScrapeIntervalSeconds = new(int32)
		*spec.ScrapeIntervalSeconds = 10
	}

	if spec.Resources == nil {
		spec.Resources = []string{"pods", "services"}
	}

	defaultedSpecOutput(spec.Output)
}

func defaultedSpecOutput(outputs *Outputs) {

}
