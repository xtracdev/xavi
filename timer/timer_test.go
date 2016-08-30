package timer

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPostitiveDuration(t *testing.T) {
	at := NewEndToEndTimer("foo")
	at.Stop(nil)
	assert.NotEqual(t, 0, at.Duration)
	assert.False(t, at.errorReported)
	assert.True(t, at.ErrorFree)
	assert.Equal(t, "", at.Error)
}

func TestTagsGetSerialized(t *testing.T) {
	at := NewEndToEndTimer("foo")
	at.Tags["foo"] = "1"
	at.Tags["bar"] = "2"
	at.Stop(nil)

	s := at.ToJSONString()
	assert.True(t, strings.Contains(s, `"bar":"2"`), "unexpected bar tag")
	assert.True(t, strings.Contains(s, `"foo":"1"`), "unexpected foo tag")
}

func TestContributors(t *testing.T) {
	at := NewEndToEndTimer("foo")
	c1 := at.StartContributor("c1")
	c2 := at.StartContributor("c2")
	time.Sleep(10 * time.Millisecond)
	c2.End(nil)
	c1.End(nil)
	at.Stop(nil)

	assert.False(t, at.errorReported)
	assert.True(t, at.ErrorFree)
	assert.Equal(t, "", at.Error)
	assert.True(t, c1.Duration > 0)
	assert.True(t, c2.Duration > 0)
}

func TestIfContributorErrorsThenTimerErrors(t *testing.T) {
	at := NewEndToEndTimer("foo")
	c1 := at.StartContributor("c1")
	c2 := at.StartContributor("c2")
	c2.End(errors.New("oh whoops"))
	c1.End(nil)
	at.Stop(nil)

	assert.False(t, at.ErrorFree)
	assert.False(t, at.errorReported)
	assert.Equal(t, "", at.Error)
}

func TestIfContributorErrorsNoErrorMessageThenTimerErrors(t *testing.T) {
	at := NewEndToEndTimer("foo")
	c1 := at.StartContributor("c1")
	c2 := at.StartContributor("c2")
	c2.End(errors.New(""))
	c1.End(nil)
	at.Stop(nil)

	assert.True(t, c2.errorReported)
	assert.False(t, c1.errorReported)
	assert.True(t, at.ContributorErrors())
	assert.False(t, at.ErrorFree)
	assert.False(t, at.errorReported)
	assert.Equal(t, "", at.Error)
}

func TestMultiBackendRecordings(t *testing.T) {
	at := NewEndToEndTimer("foo")
	c1 := at.StartContributor("c1")
	c2 := at.StartContributor("c2")
	c3 := at.StartContributor("c3")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		be1 := c3.StartServiceCall("workflo", "localhost:12345")
		be1.End(nil)
	}()

	go func() {
		defer wg.Done()
		be2 := c3.StartServiceCall("doc munger", "localhost:12345")
		be2.End(nil)
	}()

	wg.Wait()

	time.Sleep(10 * time.Millisecond)

	c3.End(nil)

	c2.End(nil)
	c1.End(nil)
	at.Stop(nil)

	assert.Equal(t, "", at.Error, "Expected error string on timer to be empty")
	assert.True(t, c1.Duration > 0, "Expected c1 Duration value > 0")
	assert.True(t, c2.Duration > 0, "Expected c2 Duration value > 0")
	assert.True(t, c3.Duration > 0, "Expected c3 Duration value > 0")
	assert.True(t, at.ErrorFree, "Expected timer error free to be true")
	assert.False(t, at.errorReported, "Expected timer error reported to be false")
	assert.Equal(t, 2, len(c3.ServiceCalls))

	println(at.ToJSONString())
}
