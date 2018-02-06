package main

import (
	"C"
	"fmt"
	"log"
	"unsafe"

	// Imports the Stackdriver Logging client package.
	"cloud.google.com/go/logging"
	"github.com/fluent/fluent-bit-go/output"
	"golang.org/x/net/context"
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

	dec := output.NewDecoder(data, int(length))

	// Iterate the original MessagePack array
	count := 0
	for {
		// Decode the entry
		ret, ts, record := output.GetRecord(dec)
		if ret != 0 {
			break
		}
		err := logger.LogSync(ctx, logging.Entry{
			Payload:   ToMarshalable(record),
			Timestamp: ts.(output.FLBTime).Time,
		})
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

func ToMarshalable(src map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range src {
		var v2 interface{}
		if w, ok := v.(map[interface{}]interface{}); ok {
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
