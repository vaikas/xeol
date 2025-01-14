package parsers

import (
	"fmt"

	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"

	"github.com/xeol-io/xeol/xeol/event"
	"github.com/xeol-io/xeol/xeol/matcher"
	"github.com/xeol-io/xeol/xeol/presenter"
)

type ErrBadPayload struct {
	Type  partybus.EventType
	Field string
	Value interface{}
}

func (e *ErrBadPayload) Error() string {
	return fmt.Sprintf("event='%s' has bad event payload field='%v': '%+v'", string(e.Type), e.Field, e.Value)
}

func newPayloadErr(t partybus.EventType, field string, value interface{}) error {
	return &ErrBadPayload{
		Type:  t,
		Field: field,
		Value: value,
	}
}

func checkEventType(actual, expected partybus.EventType) error {
	if actual != expected {
		return newPayloadErr(expected, "Type", actual)
	}
	return nil
}

func ParseAppUpdateAvailable(e partybus.Event) (string, error) {
	if err := checkEventType(e.Type, event.AppUpdateAvailable); err != nil {
		return "", err
	}

	newVersion, ok := e.Value.(string)
	if !ok {
		return "", newPayloadErr(e.Type, "Value", e.Value)
	}

	return newVersion, nil
}

func ParseEolScanningStarted(e partybus.Event) (*matcher.Monitor, error) {
	if err := checkEventType(e.Type, event.EolScanningStarted); err != nil {
		return nil, err
	}

	monitor, ok := e.Value.(matcher.Monitor)
	if !ok {
		return nil, newPayloadErr(e.Type, "Value", e.Value)
	}

	return &monitor, nil
}

func ParseEolScanningFinished(e partybus.Event) (presenter.Presenter, error) {
	if err := checkEventType(e.Type, event.EolScanningFinished); err != nil {
		return nil, err
	}

	pres, ok := e.Value.(presenter.Presenter)
	if !ok {
		return nil, newPayloadErr(e.Type, "Value", e.Value)
	}

	return pres, nil
}

func ParseNonRootCommandFinished(e partybus.Event) (*string, error) {
	if err := checkEventType(e.Type, event.NonRootCommandFinished); err != nil {
		return nil, err
	}

	result, ok := e.Value.(string)
	if !ok {
		return nil, newPayloadErr(e.Type, "Value", e.Value)
	}

	return &result, nil
}

func ParseUpdateEolDatabase(e partybus.Event) (progress.StagedProgressable, error) {
	if err := checkEventType(e.Type, event.UpdateEolDatabase); err != nil {
		return nil, err
	}

	prog, ok := e.Value.(progress.StagedProgressable)
	if !ok {
		return nil, newPayloadErr(e.Type, "Value", e.Value)
	}

	return prog, nil
}
