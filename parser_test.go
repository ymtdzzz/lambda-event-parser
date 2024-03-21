package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEvent struct {
	Message string `json:"message"`
	UserIDs []int  `json:"user_ids"`
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		path string
		want *testEvent
	}{
		{
			name: "SNS",
			path: "test/data/sns.json",
			want: &testEvent{
				Message: "test message",
				UserIDs: []int{10, 123},
			},
		},
		{
			name: "SQS",
			path: "test/data/sqs.json",
			want: &testEvent{
				Message: "test message",
				UserIDs: []int{10, 123},
			},
		},
		{
			name: "EventBridge",
			path: "test/data/eventbridge.json",
			want: &testEvent{
				Message: "test message",
				UserIDs: []int{10, 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &Event[testEvent]{}
			data, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if err := got.UnmarshalJSON(data); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want, got.Message)
		})
	}
}
