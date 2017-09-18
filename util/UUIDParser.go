package util

import (
	"github.com/cloudfoundry/sonde-go/events"
	"encoding/hex"
	"encoding/binary"
)

const dash byte = '-'

// Returns canonical string representation of UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func Stringify(uuid *events.UUID) string {
	buf := make([]byte, 36)

	lowBytes := make([]byte, 8)
	highBytes := make([]byte, 8)

	binary.LittleEndian.PutUint64(lowBytes, uuid.GetLow())
	binary.LittleEndian.PutUint64(highBytes, uuid.GetHigh())

	hex.Encode(buf[0:8], lowBytes[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], lowBytes[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], lowBytes[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], highBytes[0:2])
	buf[23] = dash
	hex.Encode(buf[24:], highBytes[2:])

	return string(buf)
}