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

package clock

import "time"

type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

type MockClock struct {
	wall string
}

func NewMockClock(ts string) *MockClock {
	return &MockClock{
		wall: ts,
	}
}

func (m *MockClock) Now() time.Time {
	now, _ := time.Parse(time.RFC3339Nano, m.wall)
	return now
}

func (m *MockClock) Since(t time.Time) time.Duration {
	now, _ := time.Parse(time.RFC3339Nano, m.wall)
	return now.Sub(t)
}

func (m *MockClock) Set(ts string) {
	m.wall = ts
}
