package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	bytes, err := ioutil.ReadFile("./test/event.json")
	assert.NoError(err)

	var request events.SQSEvent
	err = json.Unmarshal(bytes, &request)
	assert.NoError(err)

	err = handler(request)
	assert.Error(err)

	os.Setenv("REGION", "us-west-2")
	os.Setenv("RESPONSE_QUEUE", "MyQueue")
	os.Setenv("WAVEFRONT_ENABLED", "true")
	os.Setenv("WAVEFRONT_URL", "")
	os.Setenv("WAVEFRONT_API_TOKEN", "")

	err = handler(request)
	assert.NoError(err)
}
