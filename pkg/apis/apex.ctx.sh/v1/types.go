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

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BearerToken struct {
	// +optional
	String *string `json:"string,omitempty"`
	// +optional
	Path *string `json:"path,omitempty"`
}

type TLS struct {
	// +required
	CA *string `json:"ca,omitempty"`
	// +required
	Cert *string `json:"cert,omitempty"`
	// +required
	Key *string `json:"key,omitempty"`
	// +optional
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`
}

type LoggerOutput struct {
	// +required
	Enabled *bool `json:"enabled,omitempty"`
}

type DatadogOutput struct {
	// +required
	Enabled *bool `json:"enabled,omitempty"`
	// +required
	ApiKey *string `json:"apiKey,omitempty"`
	// +optional
	Timeout *time.Duration `json:"timeout,omitempty"`
	// +optional
	Url *string `json:"url,omitempty"`
	// +optional
	HttpUrlProxy *string `json:"httpUrlProxy,omitempty"`
	// +optional
	Compression *string `json:"compression,omitempty"`
}

type StatsdOutput struct {
	// +required
	Enabled *bool `json:"enabled,omitempty"`
	// +required
	Host *string `json:"host,omitempty"`
	// +optional
	Port *int32 `json:"port,omitempty"`
}

type Authentication struct {
	// +optional
	BearerToken *BearerToken `json:"bearerToken,omitempty"`
	// +optional
	Username *string `json:"username,omitempty"`
	// +optional
	Password *string `json:"password,omitempty"`
}

type Outputs struct {
	Logger  *LoggerOutput  `json:"logger,omitempty"`
	Statsd  *StatsdOutput  `json:"statsd,omitempty"`
	Datadog *DatadogOutput `json:"datadog,omitempty"`
}

type MetaTags struct {
	Name            *bool `json:"name,omitempty"`
	Namespace       *bool `json:"namespace,omitempty"`
	ResourceVersion *bool `json:"resourceVersion,omitempty"`
	Node            *bool `json:"node,omitempty"`
}

type ScraperSpec struct {
	// +optional
	Workers *int32 `json:"workers,omitempty"`
	// +optional
	AllowLabels *bool `json:"allowLabels,omitempty"`
	// +optional
	MetaTags *MetaTags `json:"metaTags,omitempty"`
	// +optional
	ScrapeIntervalSeconds *int32 `json:"scrapeIntervalSeconds,omitempty"`
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// +optional
	AnnotationPrefix *string `json:"annotationPrefix,omitempty"`
	// +optional
	Resources []string `json:"resources,omitempty"`
	// +optional
	Outputs *Outputs `json:"outputs,omitempty"`
	// ------------------------------------------------------------------------
	// These won't be implemented for the MVP, but as a follow on
	// ------------------------------------------------------------------------
	// +optional
	Authentication *Authentication `json:"authentication,omitempty"`
	// +optional
	TLS *TLS `json:"tls,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=sx,singular=scraper
// +kubebuilder:printcolumn:name="Discovered",type="string",JSONPath=".status.discovered"
// +kubebuilder:printcolumn:name="Enabled",type="string",JSONPath=".status.enabled"
// +kubebuilder:printcolumn:name="Last Scraped",type="string",JSONPath=".status.lastScraped"
type Scraper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScraperSpec `json:"spec"`
	// +optional
	Status ScraperStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ScraperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Scraper `json:"items"`
}

type ScraperStatus struct {
	Discovered  int64       `json:"discovered"`
	Enabled     int64       `json:"enabled"`
	LastScraped metav1.Time `json:"lastScraped"`
}
