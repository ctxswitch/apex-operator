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

package logger

import (
	"ctx.sh/apex-operator/pkg/metric"
	"ctx.sh/apex-operator/pkg/output"
	"github.com/go-logr/logr"
)

type Logger struct {
	log logr.Logger
}

func New(logger logr.Logger) (output.Output, error) {
	return &Logger{
		log: logger,
	}, nil
}

func (l *Logger) Send(m []metric.Metric) {
	for _, x := range m {
		l.log.Info("metric", "metric_name", x.Name(), "metric_values", x.Values(), "metric_tags", x.Tags())
	}
}

func (l *Logger) Close() {}

func (l *Logger) Name() string { return "logger" }
