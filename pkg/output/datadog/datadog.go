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

package datadog

import (
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/metric"
	"ctx.sh/apex-operator/pkg/output"
)

type Datadog struct {
	apiKey       string
	timeout      time.Duration
	url          string
	httpUrlProxy string
	compression  string
}

func New(dd apexv1.DatadogOutput) (output.Output, error) {
	return &Datadog{
		apiKey:       "",
		timeout:      time.Minute,
		url:          "",
		httpUrlProxy: "",
		compression:  "",
	}, nil
}

func (d *Datadog) Send(m []metric.Metric) {}

func (d *Datadog) Close() {}

func (d *Datadog) Name() string { return "datadog" }
