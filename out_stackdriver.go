package main
import (
	"fmt"
	"log"

	// Imports the Stackdriver Logging client package.
	"cloud.google.com/go/logging"
	"golang.org/x/net/context"
)
import "github.com/fluent/fluent-bit-go/output"
import (
	"github.com/ugorji/go/codec"
	"unsafe"
	"C"
	"reflect"
	"time"
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "stackdriver", "Stackdriver Logging")
}


// Sets your Google Cloud Platform project ID.
var projectID string

// Sets the name of the log to write to.
var logName string

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {

	// Example to retrieve an optional configuration parameter
	projectID = output.FLBPluginConfigKey(ctx, "ProjectID")
	logName = output.FLBPluginConfigKey(ctx, "LogName")
	fmt.Printf("plugin parameter ProjectID = '%s', LogName = '%s'\n", projectID, logName)
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	ctx := context.Background()

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Selects the log to write to.
	logger := client.Logger(logName)

	h := new(codec.MsgpackHandle)
	b := C.GoBytes(data, length)
	dec := codec.NewDecoderBytes(b, h)

	// Iterate the original MessagePack array
	count := 0
	for {
		// Decode the entry
		var m interface{}
		err = dec.Decode(&m)
		if err != nil {
			break
		}

		// Get a slice and their two entries: timestamp and map
		slice := reflect.ValueOf(m)
		timestamp := slice.Index(0)
		data := slice.Index(1)

		// Convert slice data to a real map and iterate
		dataMap := data.Interface().(map[interface{}] interface{})
		// Adds an entry to the log buffer.
		err = logger.LogSync(ctx, logging.Entry{
			Payload: ToMarshalable(dataMap),
			Timestamp: time.Unix(int64(timestamp.Interface().(uint64)), 0)})
		if err != nil {
			log.Fatalf("Failed to log: %v", err)
		}
		count++
	}

	// Closes the client and flushes the buffer to the Stackdriver Logging
	// service.
	if err := client.Close(); err != nil {
		log.Fatalf("Failed to close client: %v", err)
	}

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

func ToMarshalable(src map[interface{}]interface{}) map[string] interface{} {
	result := make(map[string] interface{})
	for k, v := range src {
		var v2 interface {}
		if w, ok := v.(map[interface{}] interface{}); ok {
			v2 = ToMarshalable(w)
		} else if w, ok := v.([]uint8); ok {
			v2 = string(w)
		} else {
			v2 = v
		}
		result[k.(string)] = v2
	}
	return result
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
