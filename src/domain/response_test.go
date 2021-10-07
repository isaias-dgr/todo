package domain_test

import (
	"net/url"
	"testing"

	"github.com/isaias-dgr/todo/src/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewFilter(t *testing.T) {
	val := url.Values{
		"offset":  []string{"10"},
		"limit":   []string{"10"},
		"sort_by": []string{""},
	}
	filter := domain.NewFilter(val)
	assert.Equal(t, filter.Limit, 10)
}

func TestNewFilterError(t *testing.T) {
	val := url.Values{
		"offset":  []string{"a"},
		"limit":   []string{"10"},
		"sort_by": []string{""},
	}
	filter := domain.NewFilter(val)
	assert.Equal(t, filter.Offset, 0)
}
func TestNewMetadata(t *testing.T) {
	val := url.Values{
		"offset":  []string{"10"},
		"limit":   []string{"10"},
		"sort_by": []string{""},
	}
	filter := domain.NewFilter(val)
	meta := domain.NewMetadata(10, filter, "success")
	assert.Equal(t, meta.Message, "success")
	assert.Equal(t, meta.Offset, 10)
}

func TestNewResponse(t *testing.T) {
	val := url.Values{
		"offset":  []string{"10"},
		"limit":   []string{"10"},
		"sort_by": []string{""},
	}
	filter := domain.NewFilter(val)
	data := map[string]string{
		"key 1": "value 1",
		"key 2": "value 2",
		"key 3": "value 3",
	}
	resp := domain.NewResponse(data, 10, filter, "success")
	assert.Equal(t, resp.Data, data)
}
