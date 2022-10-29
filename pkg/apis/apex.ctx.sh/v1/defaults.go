/*
 * Copyright 2022 Rob Lyon <rob@ctxswitch.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
