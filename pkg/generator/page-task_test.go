package generator

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchSource_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	previousTimeout := sourceRequestTimeout
	sourceRequestTimeout = 50 * time.Millisecond
	defer func() {
		sourceRequestTimeout = previousTimeout
	}()

	_, err := fetchSource(DataTask{Key: "slow", UrlTmpl: server.URL}, nil)
	assert.Error(t, err)
}
