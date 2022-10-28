package v1

// Defaulted sets the scraper resource defaults
func Defaulted(scraper *Scraper) {
	defaultedSpec(&scraper.Spec)
}

func defaultedSpec(spec *ScraperSpec) {
	if spec.Workers == nil {
		spec.Workers = new(int32)
		*spec.Workers = 10
	}

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

	if spec.AllowLabels == nil {
		spec.AllowLabels = new(bool)
		*spec.AllowLabels = false
	}

	defaultedSpecMetaTags(spec.MetaTags)
	defaultedSpecOutput(spec.Outputs)
}

func defaultedSpecMetaTags(m *MetaTags) {
	if m == nil {
		m = &MetaTags{}
	}

	if m.Name == nil {
		m.Name = new(bool)
		*m.Name = false
	}

	if m.Namespace == nil {
		m.Namespace = new(bool)
		*m.Namespace = false
	}

	if m.ResourceVersion == nil {
		m.ResourceVersion = new(bool)
		*m.ResourceVersion = false
	}

	if m.Node == nil {
		m.Node = new(bool)
		*m.Node = false
	}
}

func defaultedSpecOutput(outputs *Outputs) {
	if outputs == nil {
		logger := &LoggerOutput{}
		logger.Enabled = new(bool)
		*logger.Enabled = true

		outputs = &Outputs{
			Logger: logger,
		}
		return
	}

	defaultedSpecOutputStatsd(outputs.Statsd)
	defaultedSpecOutputDatadog(outputs.Datadog)
	defaultedSpecOutputLogger(outputs.Logger)
}

func defaultedSpecOutputStatsd(o *StatsdOutput) {
	if o == nil {
		return
	}

	// Host is required

	if o.Port == nil {
		o.Port = new(int32)
		*o.Port = 8125
	}
}

func defaultedSpecOutputDatadog(o *DatadogOutput) {
	if o == nil {
		return
	}
}

func defaultedSpecOutputLogger(o *LoggerOutput) {}
