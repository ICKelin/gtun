package god

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBSet(t *testing.T) {
	type TestCase struct {
		key interface{}
		val interface{}
	}

	var testCases = []TestCase{
		{
			key: "123",
			val: 1,
		},
		{
			key: "234",
			val: 2,
		},
	}

	db := NewDB()

	for _, testCase := range testCases {
		db.Set(testCase.key, testCase.val)
		v, ok := db.Get(testCase.key)
		assert.Equal(t, true, ok)
		assert.Equal(t, testCase.val, v)
	}

	val, ok := db.Get("not insert key")
	assert.Equal(t, false, ok)
	assert.Equal(t, nil, val)
}
