package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Authentication struct {
	// +optional
	BearerToken *BearerToken `json:"bearerToken,omitempty"`
	// +optional
	Username *string `json:"username,omitempty"`
	// +optional
	Password *string `json:"password,omitempty"`
}

type BearerToken struct {
	// +optional
	String *string `json:"string,omitempty"`
	// +optional
	Path *string `json:"path,omitempty"`
}

type TLS struct {
	// +required
	CA string `json:"ca,omitempty"`
	// +required
	Cert string `json:"cert,omitempty"`
	// +required
	Key string `json:"key,omitempty"`
	// +optional
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`
}

type ScraperSpec struct {
	// +required
	Name string `json:"name"`
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// +optional
	ScrapeInterval *int32 `json:"scrapeInterval,omitempty"`
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// ------------------------------------------------------------------------
	// These won't be implemented for the MVP, but as a follow on
	// ------------------------------------------------------------------------
	// +optional
	Authentication Authentication `json:"authentication,omitempty"`
	// +optional
	TLS *TLS `json:"tls,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=sx,singular=scraper
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
// +kubebuilder:printcolumn:name="Pods",type="string",JSONPath=".status.totalPods"
// +kubebuilder:printcolumn:name="Services",type="string",JSONPath=".status.totalServices"
// +kubebuilder:printcolumn:name="Errors (pods)",type="string",JSONPath=".status.erroredPods"
// +kubebuilder:printcolumn:name="Errors (services)",type="string",JSONPath=".status.erroredServices"
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get
// +kubebuilder:rbac:groups=apex.ctx.sh,resources=scraper,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apex.ctx.sh,resources=scraper/status,verbs=get;update;patch
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
	Version       string `json:"version"`
	TotalPods     int64  `json:"totalPods"`
	TotalServices int64  `json:"totalServices"`
	OkPods        int64  `json:"okPods"`
	OkServices    int64  `json:"okServices"`
	ErrorPods     int64  `json:"errorPods"`
	ErrorServices int64  `json:"errorServices"`
}
