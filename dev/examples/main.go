// Copyright (C) 2022, Rob Lyon <rob@ctxswitch.com>
//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package main

import (
	"context"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"ctx.sh/apex"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var (
	histogramOpts = apex.HistogramOpts{
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5},
	}
	summaryOpts = apex.SummaryOpts{
		MaxAge:     10 * time.Minute,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		AgeBuckets: 5,
	}
)

type Handlers struct {
	logger  *zap.Logger
	metrics *apex.Metrics
}

func NewHandlers(logger *zap.Logger, metrics *apex.Metrics) *Handlers {
	return &Handlers{logger: logger, metrics: metrics}
}

func (h *Handlers) DefaultHandler() http.HandlerFunc {
	labels := apex.Labels{
		"func":   "DefaultHandler",
		"region": "us-east-1",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		timer := h.metrics.HistogramTimer("latency", labels, histogramOpts)
		defer timer.ObserveDuration()

		h.logger.Info("request recieved", zap.String("uri", r.RequestURI), zap.String("method", r.Method))
		defer r.Body.Close()

		h.metrics.SummaryObserve("test_summary", random(0, 10), labels, summaryOpts)

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) Observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labels := apex.Labels{
			"region": "us-east-1",
		}
		m := httpsnoop.CaptureMetrics(next, w, r)
		h.metrics.GaugeSet("response_code", float64(m.Code), labels)
		h.metrics.GaugeSet("response_size", float64(m.Written), labels)
		h.metrics.SummaryObserve("response_duration", float64(m.Duration), labels, summaryOpts)

	})
}

func random(min int, max int) float64 {
	return float64(min) + rand.Float64()*(float64(max-min))
}

func main() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger.Info("starting server")

	metrics := apex.New(apex.MetricsOpts{
		Namespace:    "apex",
		Subsystem:    "example",
		Separator:    ':',
		PanicOnError: true,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = metrics.Start()
	}()

	hnd := NewHandlers(logger, metrics)

	router := mux.NewRouter()
	router.HandleFunc("/", hnd.DefaultHandler()).Methods(http.MethodGet)
	router.NotFoundHandler = hnd.DefaultHandler()
	router.MethodNotAllowedHandler = hnd.DefaultHandler()
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server exited", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Info("server shutting down")
	shutdown(server)

	wg.Wait()
	logger.Info("server shut down successfully")
}

func shutdown(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_ = srv.Shutdown(ctx)
}
