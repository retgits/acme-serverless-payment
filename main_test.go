package main

import (
	"encoding/json"
	"io/ioutil"
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
	assert.NoError(err)

	bytes, err = ioutil.ReadFile("./test/failure.json")
	assert.NoError(err)

	err = json.Unmarshal(bytes, &request)
	assert.NoError(err)

	err = handler(request)
	assert.NoError(err)
}
