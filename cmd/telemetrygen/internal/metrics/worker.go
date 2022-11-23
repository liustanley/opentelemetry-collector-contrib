// Copyright The OpenTelemetry Authors
// Copyright (c) 2018 The Jaeger Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric/instrument"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type worker struct {
	running       *atomic.Bool    // pointer to shared flag that indicates it's time to stop the test
	totalDuration time.Duration   // how long to run the test for
	wg            *sync.WaitGroup // notify when done
	metricType    string          // type of metric to be generated
	logger        *zap.Logger
}

func (w worker) simulateMetrics() {
	for w.running.Load() {
		meter := global.MeterProvider().Meter("telemetrygen")
		// labels represent additional key-value descriptors that can be bound to a
		// metric observer or recorder.
		commonLabels := []attribute.KeyValue{
			attribute.String("labelA", "chocolate"),
			attribute.String("labelB", "raspberry"),
			attribute.String("labelC", "vanilla"),
		}

		if w.metricType == "gauge" {
			gauge, err := meter.AsyncFloat64().Gauge(
				"telemetrygen.gauge",
				instrument.WithUnit("1"),
				instrument.WithDescription("Test gauge"),
			)
			if err != nil {
				w.logger.Fatal("failed to create gauge", zap.Error(err))
			}

			if err = meter.RegisterCallback(
				[]instrument.Asynchronous{
					gauge,
				},
				func(ctx context.Context) {
					gauge.Observe(ctx, 5.0, commonLabels...)
				},
			); err != nil {
				w.logger.Fatal("failed to register gauge callback", zap.Error(err))
			}

			// work begins
			for i := 0; i < 20; i++ {
				<-time.After(time.Second)
			}

			// Give a bit of time for the exporter to process the last metrics
			<-time.After(10 * time.Second)
		} else if w.metricType == "counter" {
			counter, err := meter.SyncInt64().Counter(
				"telemetrygen.counter",
				instrument.WithUnit("1"),
				instrument.WithDescription("Test counter"),
			)
			if err != nil {
				w.logger.Fatal("failed to create counter", zap.Error(err))
			}

			// work begins
			for i := 0; i < 20; i++ {
				counter.Add(context.Background(), 1, commonLabels...)
				<-time.After(time.Second)
			}

			// Give a bit of time for the exporter to process the last metrics
			<-time.After(10 * time.Second)
		} else if w.metricType == "histogram" {
			histogram, err := meter.SyncFloat64().Histogram(
				"telemetrygen.histogram",
				instrument.WithUnit("1"),
				instrument.WithDescription("Test histogram"),
			)
			if err != nil {
				w.logger.Fatal("failed to create histogram", zap.Error(err))
			}

			// work begins
			for i := 0; i < 20; i++ {
				histogram.Record(context.Background(), float64(i+1), commonLabels...)
				<-time.After(time.Second)
			}

			// Give a bit of time for the exporter to process the last metrics
			<-time.After(10 * time.Second)
		}

		break
	}
	w.logger.Info("metrics generated")
	w.wg.Done()
}
