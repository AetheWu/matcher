package matcher

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func newMqttMatcherDemo() Matcher[struct{}] {
	matcher := NewMqttTopicMatcher[struct{}]()
	matcher.AddPath("iot/bms/things/+/up/props")
	matcher.AddPath("iot/bms/things/+/up/props/bulk")
	matcher.AddPath("iot/bms/things/+/up/raw-data")
	matcher.AddPath("iot/bms/things/+/up/service")
	matcher.AddPath("iot/bms/things/+/up/ota/+")
	matcher.AddPath("iot/bms/things/+/up/event")
	return matcher
}

func newNatsMatcherDemo() Matcher[struct{}] {
	matcher := NewNatsSubjectMatcher[struct{}]()
	matcher.AddPath("iot.bms.things.prop.>.>")
	matcher.AddPath("iot.bms.things.propBulk.>.>")
	matcher.AddPath("iot.bms.things.propHist.>.>")
	matcher.AddPath("iot.bms.things.propHistBulk.>.>")
	return matcher
}

func Benchmark_Match(b *testing.B) {
	b.Run("mqtt topic match", func(b *testing.B) {
		matcher := newMqttMatcherDemo()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			matcher.MatchWithAnonymousParams(fmt.Sprintf("iot/bms/things/%d/up/props/bulk", i))
		}
	})

	b.Run("mqtt topic match", func(b *testing.B) {
		matcher := newMqttMatcherDemo()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			matcher.Match(fmt.Sprintf("iot/bms/things/%d/up/props/bulk", i))
		}
	})
}

func TestMatcher_MqttMatch(t *testing.T) {
	tr := newMqttMatcherDemo()
	type args struct {
		dstTopic string
	}
	tests := []struct {
		name            string
		args            args
		wantMatchedPath string
		wantParams      []string
		wantOk          bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				dstTopic: "iot/bms/things/hello/up/props",
			},
			wantMatchedPath: "iot/bms/things/+/up/props",
			wantParams:      []string{"hello"},
			wantOk:          true,
		},
		{
			name: "test1",
			args: args{
				dstTopic: "iot/bms/things/sxx_heloo/up/raw-data",
			},
			wantMatchedPath: "iot/bms/things/+/up/raw-data",
			wantParams:      []string{"sxx_heloo"},
			wantOk:          true,
		},
		{
			name: "not match",
			args: args{
				dstTopic: "iot/bms/things/sxx_heloo/up/x",
			},
			wantMatchedPath: "",
			wantParams:      nil,
			wantOk:          false,
		},
		{
			name: "ota",
			args: args{
				dstTopic: "iot/bms/things/edge1/up/ota/upgradePost",
			},
			wantMatchedPath: "iot/bms/things/+/up/ota/+",
			wantParams:      []string{"edge1", "upgradePost"},
			wantOk:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatchedPath, gotParams, gotOk := tr.MatchWithAnonymousParams(tt.args.dstTopic)
			if gotOk != tt.wantOk {
				t.Errorf("Matcher.MatchWithAnonymousParams() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk {
				if gotMatchedPath != tt.wantMatchedPath {
					t.Errorf("Matcher.MatchWithAnonymousParams() gotMatchedPath = %v, want %v", gotMatchedPath, tt.wantMatchedPath)
				}
				if !reflect.DeepEqual(gotParams, tt.wantParams) {
					t.Errorf("Matcher.MatchWithAnonymousParams() gotParams = %v, want %v", gotParams, tt.wantParams)
				}
			}
		})
	}
}

func TestMatchWithAnonymousParamsAndValues(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*Matcher[string])
		matchPath      string
		expectedParams []string
		expectedValues []string
		expectedOk     bool
	}{
		{
			name: "ExactMatch",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/home", "home_handler")
			},
			matchPath:      "/home",
			expectedParams: []string{},
			expectedValues: []string{"home_handler"},
			expectedOk:     true,
		},
		{
			name: "ParamMatch",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/user/:id", "user_handler")
			},
			matchPath:      "/user/123",
			expectedParams: []string{"123"},
			expectedValues: []string{"user_handler"},
			expectedOk:     true,
		},
		{
			name: "WildcardMatch",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/api/*", "api_handler")
			},
			matchPath:      "/api/v1/resource",
			expectedParams: []string{},
			expectedValues: []string{"api_handler"},
			expectedOk:     true,
		},
		{
			name: "MultiParamMatch",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/:year/:month/:day", "date_handler")
			},
			matchPath:      "/2023/04/05",
			expectedParams: []string{"2023", "04", "05"},
			expectedValues: []string{"date_handler"},
			expectedOk:     true,
		},
		{
			name: "PriorityMatch",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/user/admin", "admin_handler")
				m.AddPathWithValue("/user/:id", "user_handler")
			},
			matchPath:      "/user/admin",
			expectedParams: []string{},
			expectedValues: []string{"admin_handler"},
			expectedOk:     true,
		},
		{
			name: "NoMatch",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/home", "home_handler")
			},
			matchPath:      "/about",
			expectedParams: []string{},
			expectedValues: nil,
			expectedOk:     false,
		},
		{
			name: "MultiValueSamePath",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/multi", "handler1")
				m.AddPathWithValue("/multi", "handler2")
			},
			matchPath:      "/multi",
			expectedParams: []string{},
			expectedValues: []string{"handler1", "handler2"},
			expectedOk:     true,
		},
		{
			name: "MixedParamWildcard",
			setup: func(m *Matcher[string]) {
				m.AddPathWithValue("/:version/api/*", "versioned_api")
			},
			matchPath:      "/v1/api/resource/123",
			expectedParams: []string{"v1"},
			expectedValues: []string{"versioned_api"},
			expectedOk:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRouterPathMatcher[string]()
			tt.setup(&m)

			matchedPath, params, values, ok := m.MatchWithAnonymousParamsAndValues(tt.matchPath)

			if tt.expectedOk {
				if matchedPath == "" {
					t.Errorf("Expected matched path, got empty")
				}
			} else {
				if matchedPath != "" {
					t.Errorf("case[%s] Expected no match, got path: %s", tt.name, matchedPath)
				}
			}

			if !reflect.DeepEqual(params, tt.expectedParams) {
				t.Errorf("case[%s] Params mismatch: got %v, want %v", tt.name, params, tt.expectedParams)
			}

			if !reflect.DeepEqual(values, tt.expectedValues) {
				t.Errorf("case[%s] Values mismatch: got %v, want %v", tt.name, values, tt.expectedValues)
			}

			if ok != tt.expectedOk {
				t.Errorf("case[%s] Match flag mismatch: got %v, want %v", tt.name, ok, tt.expectedOk)
			}
		})
	}
}

func FuzzMatchWithAnonymousParams(f *testing.F) {
	m := NewRouterPathMatcher[string]()
	m.AddPathWithValue("/:version/api/*", "versioned_api")

	f.Add("/v1/api/resource/123")
	f.Add("/v2/api/test")

	f.Fuzz(func(t *testing.T, path string) {
		_, _, _, ok := m.MatchWithAnonymousParamsAndValues(path)
		// 验证不会panic并返回合理结果
		if ok && !strings.HasPrefix(path, "/") {
			t.Errorf("Unexpected match for invalid path: %s", path)
		}
	})
}
