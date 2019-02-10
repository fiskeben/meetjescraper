package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fiskeben/scrapejestad"
)

func Test_server(t *testing.T) {
	type args struct {
		sensor string
		limit  string
	}
	type want struct {
		status      int
		numReadings int
		contentType string
		body        string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "returns 200 ok",
			args: args{
				sensor: "242",
				limit:  "5",
			},
			want: want{
				status:      http.StatusOK,
				numReadings: 5,
				contentType: "application/json",
			},
		},
		{
			name: "defaults to 50 elements",
			args: args{
				sensor: "242",
			},
			want: want{
				status:      http.StatusOK,
				numReadings: 50,
				contentType: "application/json",
			},
		},
		{
			name: "returns 400 if sensor ID is missing",
			args: args{
				limit: "5",
			},
			want: want{
				status:      400,
				contentType: "application/json",
			},
		},
		{
			name: "returns 400 if sensor ID is not a number",
			args: args{
				sensor: "three",
			},
			want: want{
				status:      400,
				contentType: "application/json",
			},
		},
		{
			name: "returns 400 if limit is not a number",
			args: args{
				limit: "yes",
			},
			want: want{
				status:      400,
				contentType: "application/json",
			},
		},
		{
			name: "returns 400 if limit is higher than 100",
			args: args{
				limit:  "101",
				sensor: "242",
			},
			want: want{
				status:      400,
				contentType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/?sensor=%s&limit=%s", tt.args.sensor, tt.args.limit), nil)
		if err != nil {
			t.Fatalf("error creating request: %v", err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handle)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.want.status {
			t.Errorf("expected %d got %d", tt.want.status, status)
		}

		if contentType := rr.Header().Get("content-type"); contentType != tt.want.contentType {
			t.Errorf("expected content type '%s' got '%s'", tt.want.contentType, contentType)
		}

		if tt.want.status == http.StatusOK {
			decoder := json.NewDecoder(rr.Body)
			var data []scrapejestad.Reading
			if err := decoder.Decode(&data); err != nil {
				t.Errorf("failed to parse body: %v", err)
			}
			if len(data) != tt.want.numReadings {
				t.Errorf("expected %d entries got %d", tt.want.numReadings, len(data))
			}
		}

	}

}
