package lora

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/mainflux/mainflux"
)

const (
	protocol      = "lora"
	thingSuffix   = "thing"
	channelSuffix = "channel"
)

var (
	// ErrMalformedIdentity indicates malformed identity received (e.g.
	// invalid appID or deviceEUI).
	ErrMalformedIdentity = errors.New("malformed identity received")

	// ErrMalformedMessage indicates malformed LoRa message.
	ErrMalformedMessage = errors.New("malformed message received")

	// ErrNotFoundDev indicates a non-existent route map for a device EUI.
	ErrNotFoundDev = errors.New("route map not found for this device EUI")

	// ErrNotFoundApp indicates a non-existent route map for an application ID.
	ErrNotFoundApp = errors.New("route map not found for this application ID")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateThing creates thing  mfx:lora & lora:mfx route-map
	CreateThing(string, string) error

	// UpdateThing updates thing mfx:lora & lora:mfx route-map
	UpdateThing(string, string) error

	// RemoveThing removes thing mfx:lora & lora:mfx route-map
	RemoveThing(string) error

	// CreateChannel creates channel mfx:lora & lora:mfx route-map
	CreateChannel(string, string) error

	// UpdateChannel updates mfx:lora & lora:mfx route-map
	UpdateChannel(string, string) error

	// RemoveChannel removes channel mfx:lora & lora:mfx route-map
	RemoveChannel(string) error

	// Publish forwards messages from the LoRa MQTT broker to Mainflux NATS broker
	Publish(context.Context, string, Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher  mainflux.MessagePublisher
	thingsRM   RouteMapRepository
	channelsRM RouteMapRepository
}

// New instantiates the LoRa adapter implementation.
func New(pub mainflux.MessagePublisher, thingsRM, channelsRM RouteMapRepository) Service {
	return &adapterService{
		publisher:  pub,
		thingsRM:   thingsRM,
		channelsRM: channelsRM,
	}
}

func FloatToString(input_num float64) string {
    // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', 1, 64)
}

func IntToString(input_num int64) string {
    // to convert an int number to a string
    return strconv.FormatInt(int64(input_num), 10)
}


// Publish forwards messages from Lora MQTT broker to Mainflux NATS broker
func (as *adapterService) Publish(ctx context.Context, token string, m Message) error {
	// Get route map of lora application
	thing, err := as.thingsRM.Get(m.DevEUI)
	if err != nil {
		return ErrNotFoundDev
	}

	// Get route map of lora application
	channel, err := as.channelsRM.Get(m.ApplicationID)
	if err != nil {
		return ErrNotFoundApp
	}

	// Use the SenML message decoded on LoRa server application if
	// field Object isn't empty. Otherwise, decode standard field Data.
	var payload []byte
	switch m.Object {
	case nil:
		payload, err = base64.StdEncoding.DecodeString(m.Data)
		if err != nil {
			return ErrMalformedMessage
		}
	default:
		jo, err := json.Marshal(m.Object)
		if err != nil {
			return err
		}

		tmp := string(jo)
		tmp = tmp[:len(tmp)-1]

		if len(m.RxInfo) > 0 {
			tmp = tmp + ",{\"n\":\"rssi\",\"u\":\"dBm\",\"v\":" + IntToString(int64(m.RxInfo[0].Rssi)) + "}"
		    tmp = tmp + ",{\"n\":\"snr\",\"u\":\"dB\",\"v\":" + FloatToString(m.RxInfo[0].LoRaSNR) + "}"
		}

		tmp = tmp + ",{\"n\":\"dr\",\"v\":" + IntToString(int64(m.TxInfo.Dr)) + "}"
		tmp = tmp + ",{\"n\":\"freq\",\"v\":" + IntToString(int64(m.TxInfo.Frequency)) + "}"
		tmp = tmp + "]"

		payload = []byte(tmp)
	}

	// Publish on Mainflux NATS broker
	msg := mainflux.Message{
		Publisher:   thing,
		Protocol:    protocol,
		//ContentType: "application_json",
		Channel:     channel,
		Payload:     payload,
		Subtopic:    thing,
	}

	return as.publisher.Publish(ctx, token, msg)
}

func (as *adapterService) CreateThing(mfxDevID string, loraDevEUI string) error {
	return as.thingsRM.Save(mfxDevID, loraDevEUI)
}

func (as *adapterService) UpdateThing(mfxDevID string, loraDevEUI string) error {
	return as.thingsRM.Save(mfxDevID, loraDevEUI)
}

func (as *adapterService) RemoveThing(mfxDevID string) error {
	return as.thingsRM.Remove(mfxDevID)
}

func (as *adapterService) CreateChannel(mfxChanID string, loraAppID string) error {
	return as.channelsRM.Save(mfxChanID, loraAppID)
}

func (as *adapterService) UpdateChannel(mfxChanID string, loraAppID string) error {
	return as.channelsRM.Save(mfxChanID, loraAppID)
}

func (as *adapterService) RemoveChannel(mfxChanID string) error {
	return as.channelsRM.Remove(mfxChanID)
}
