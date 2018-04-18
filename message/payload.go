package message

import (
	"fmt"
	"reflect"

	"github.com/golemfactory/bootstrap_go/cbor"
)

// slot is a pair of python field's name and value
type messageSlot = []interface{}

// list of messageSlots
type messagePayload = []interface{}

func getSerializedPayload(msg Message) ([]byte, error) {
	payload := messagePayload{}
	v := reflect.ValueOf(msg).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		val := v.Field(i)
		tag := field.Tag.Get("msg_slot")
		if tag != "" {
			payload = append(payload, messageSlot{tag, val.Interface()})
		}
	}
	return cbor.Serialize(payload)
}

func deserializePayload(rawPayload []byte, msg Message) error {
	var maybeSlots interface{}
	err := cbor.Deserialize(rawPayload, &maybeSlots)
	if err != nil {
		return err
	}
	slotsList, ok := maybeSlots.(messagePayload)
	if !ok {
		return fmt.Errorf("incorrect format of message payload")
	}

	slots := make(map[string]interface{})
	for _, s := range slotsList {
		slot, ok := s.(messageSlot)
		if !ok {
			fmt.Printf("Couldn't cast slot %+v\n", s)
			continue
		}
		if len(slot) != 2 {
			fmt.Printf("Slot should be of length 2, got %+v\n", slot)
			continue
		}
		if slotName, ok := slot[0].(string); ok {
			slots[slotName] = slot[1]
		} else {
			fmt.Printf("Expected slot name to be a string, got %+v", slot[0])
		}
	}

	v := reflect.ValueOf(msg).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		val := v.Field(i)
		tag := field.Tag.Get("msg_slot")
		if tag != "" {
			if vv, ok := slots[tag]; ok && vv != nil {
				val.Set(reflect.ValueOf(vv))
			}
		}
	}

	return nil
}
