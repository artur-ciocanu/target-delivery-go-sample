package visitor

import (
	"fmt"
	"math/rand"
	"time"
)

type SupplementalDataIdState struct {
	Current         string `json:"supplementalDataIDCurrent"`
	CurrentConsumed map[string]bool `json:"supplementalDataIDCurrentConsumed"`
	Last            string `json:"supplementalDataIDLast"`
	LastConsumed    map[string]bool `json:"supplementalDataIDLastConsumed"`
}

var chars = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}
var state = &SupplementalDataIdState{
	Current:         "",
	CurrentConsumed: make(map[string]bool),
	Last:            "",
	LastConsumed:    make(map[string]bool),
}

func generateId() string {
	top := make([]byte, 16)
	bottom := make([]byte, 16)
	maxIndex := 8

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 16; i++ {
		top[i] = chars[rand.Int()%maxIndex]
		bottom[i] = chars[rand.Int()%maxIndex]
		maxIndex = 16
	}

	return fmt.Sprintf("%s-%s", string(top), string(bottom))
}

func GetSupplementalDataId(consumerId string) string {
	if state.Current == "" {
		state.Current = generateId()
	}

	// Default to using the current supplemental-data ID
	var supplementalDataId = state.Current

	// If we have the last supplemental-data ID that has not been consumed by this consumer...
	if state.Last != "" && !state.LastConsumed[consumerId] {
		// Use the last supplemental-data ID
		supplementalDataId = state.Last
		// Mark the last supplemental-data ID as consumed for this consumer
		state.LastConsumed[consumerId] = true
		// If we are using te current supplemental-data ID at this point and we have a supplemental-data ID...
	} else if supplementalDataId != "" {
		// If the current supplemental-data ID has already been consumed by this consumer..
		if state.CurrentConsumed[consumerId] {
			supplementalDataId = generateId()
			// Move the current supplemental-data ID to the last including the current consumed list
			state.Last = state.Current
			state.LastConsumed = state.CurrentConsumed
			// Generate a new current supplemental-data ID if needed, use it, and clear the current consumed list
			state.Current = supplementalDataId
			state.CurrentConsumed = make(map[string]bool)
		}
		// If we still have a supplemental-data ID mark the current supplemental-data ID as consumed by this consumer
		if supplementalDataId != "" {
			state.CurrentConsumed[consumerId] = true
		}
	}

	return supplementalDataId
}

func GetState(orgId string) map[string]*SupplementalDataIdState {
	result := make(map[string]*SupplementalDataIdState)
	result[orgId] = state

	return result
}
