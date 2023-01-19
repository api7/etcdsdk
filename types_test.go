package etcdsdk

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDMarshalJSON(t *testing.T) {
	b := BaseInfo{
		ID: ID("123"),
	}
	marshal, err := json.Marshal(b)
	assert.Nil(t, err, "checking marshal json")
	assert.Equal(t, `{"id":"123"}`, string(marshal), "checking ID marshal json body")
}

func TestIDUnmarshalJSON(t *testing.T) {
	jsonStr := `{"id":"123","create_time":11,"update_time":22}`
	marshal := bytes.NewBufferString(jsonStr).Bytes()

	var b BaseInfo
	err := json.Unmarshal(marshal, &b)
	assert.Nil(t, err, "checking unmarshal json")
	assert.Equal(t, b.ID, ID("123"), "checking ID")
	assert.Equal(t, b.UpdateTime, int64(22), "checking name")

	jsonStr = `{"id":123,"create_time":11,"update_time":22}`
	marshal = bytes.NewBufferString(jsonStr).Bytes()
	err = json.Unmarshal(marshal, &b)
	assert.Nil(t, err, "checking unmarshal json")
	assert.Equal(t, b.ID, ID("123"), "checking ID")
	assert.Equal(t, b.UpdateTime, int64(22), "checking name")
}
