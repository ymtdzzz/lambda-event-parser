package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

// Event is incoming event
type Event[T any] struct {
	Message *T
}

type eventType int

const (
	unknownEventType eventType = iota // lamdaのコンソールからのテストもこのイベントタイプに該当する
	sqsEventType
	snsEventType
	eventBridgeEventType
)

func (event *Event[T]) getEventType(data []byte) (eventType, error) {
	temp := make(map[string]interface{})
	if err := json.Unmarshal(data, &temp); err != nil {
		return unknownEventType, err
	}

	if source, ok := temp["source"].(string); ok && source == "aws.events" {
		return eventBridgeEventType, nil
	}
	recordsList, _ := temp["Records"].([]interface{})
	record, _ := recordsList[0].(map[string]interface{})

	var eventSource string

	if es, ok := record["EventSource"]; ok {
		eventSource = es.(string)
	} else if es, ok := record["eventSource"]; ok {
		eventSource = es.(string)
	}

	switch eventSource {
	case "aws:sqs":
		return sqsEventType, nil
	case "aws:sns":
		return snsEventType, nil
	}

	return unknownEventType, nil
}

func (e *Event[T]) UnmarshalJSON(v []byte) error {
	et, err := e.getEventType(v)
	if err != nil {
		return err
	}
	switch et {
	case sqsEventType:
		sqsEvent := &events.SQSEvent{}
		err := json.Unmarshal(v, sqsEvent)
		if err != nil && len(sqsEvent.Records) == 0 {
			return errors.Wrap(err, "failed to unmarshal sqs event")
		}
		var msg T
		err = json.Unmarshal([]byte(sqsEvent.Records[0].Body), &msg)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal sqs event body")
		}
		e.Message = &msg
		return nil

	case snsEventType:
		snsEvent := &events.SNSEvent{}
		err = json.Unmarshal(v, snsEvent)
		if err != nil && len(snsEvent.Records) == 0 {
			return errors.Wrap(err, "failed to unmarshal sns event")
		}
		var msg T
		err = json.Unmarshal([]byte(snsEvent.Records[0].SNS.Message), &msg)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal sns event body")
		}
		e.Message = &msg
		return nil

	case eventBridgeEventType:
		eventBridgeEvent := &events.EventBridgeEvent{}
		err := json.Unmarshal(v, eventBridgeEvent)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal event bridge event")
		}
		var msg T
		err = json.Unmarshal([]byte(eventBridgeEvent.Detail), &msg)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal event bridge event body")
		}
		e.Message = &msg
		return nil

	case unknownEventType:
		fmt.Printf("unknown event type: %s\n", string(v))
		return nil
	}

	return nil
}
