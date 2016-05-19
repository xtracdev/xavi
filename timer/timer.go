/*
Package timer implements a simple timing utility that can be used to capture
end to end timings, plus any subtimings of interests.
*/
package timer

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

//ServiceCall is used to capture a service call made in the context of a Contributor timing.
type ServiceCall struct {
	sync.RWMutex
	Name          string
	Endpoint      string
	Duration      time.Duration
	Error         string
	errorReported bool
	start         time.Time
}

//Contributor is used to capture sub-timings of note that contribute to the end to end time.
type Contributor struct {
	sync.RWMutex
	Name          string
	Duration      time.Duration
	Error         string
	errorReported bool
	start         time.Time
	ServiceCalls  []*ServiceCall
}

//EndToEndTimer is used to capture an end to end timing. Subtimings can be added to
//an end to end time using StartContributor. Note logging timestamp is serialized as
//time so we can have a single logging time index in elasticsearch
type EndToEndTimer struct {
	sync.RWMutex
	Name             string
	Tags             map[string]string
	Duration         time.Duration
	LoggingTimestamp time.Time `json:"time"`
	TxnId            string
	Contributors     []*Contributor
	ErrorFree        bool
	Error            string
	errorReported    bool
	start            time.Time
}

//NewEndToEndTimer creates a new EndToEndTimer. Note that the clock starts when an
//EndToEndTimer is created.
func NewEndToEndTimer(name string) *EndToEndTimer {
	return &EndToEndTimer{
		TxnId: makeTxnId(),
		Name:  name,
		start: time.Now(),
		Tags:  make(map[string]string),
	}
}

//Stop stops the clock for an EndToEndTimer. If an error occurred in the timing path,
//it should be noted by passing an error object to Stop, otherwise pass nil. The JSON
//representation of the timing data will reflect if an error occurred during the timing.
func (t *EndToEndTimer) Stop(err error) {
	stopTime := time.Now()
	contribErrors := t.ContributorErrors()

	t.Lock()

	t.Duration = stopTime.Sub(t.start)
	if err != nil {
		t.Error = err.Error()
		t.errorReported = true
	}
	t.ErrorFree = contribErrors == false && t.errorReported == false

	t.Unlock()
}

//StartContributor creates a Contributor for capturing a sub timing,
//starting the clock on the subtimer.
func (t *EndToEndTimer) StartContributor(name string) *Contributor {
	contributor := &Contributor{
		Name:  name,
		start: time.Now(),
	}

	t.Lock()
	t.Contributors = append(t.Contributors, contributor)
	t.Unlock()

	return contributor
}

//ContributorErrors returns true if any subtiming has been stopped
//with an error indication
func (t *EndToEndTimer) ContributorErrors() bool {
	var foundError bool
	t.RLock()
	for _, c := range t.Contributors {
		if c.errorReported {
			foundError = true
			break
		}
	}
	t.RUnlock()
	return foundError
}

//ToJSONString produces a JSON string representation of an EndToEndTimer, including all of
//its contributors.
func (t *EndToEndTimer) ToJSONString() string {
	t.RLock()
	defer t.RUnlock()
	for _, c := range t.Contributors {
		c.RLock()
		defer c.RUnlock()
		for _, sc := range c.ServiceCalls {
			sc.RLock()
			defer sc.RUnlock()
		}
	}

	s, err := json.Marshal(t)
	if err != nil {
		s = []byte("{}")
	}
	return string(s)
}

//End stops the clock for a contributor. Any errors of note that occur
//during the contributor should be pass along in the err argument, otherwise
//pass nil.
func (c *Contributor) End(err error) {
	c.Lock()
	c.Duration = time.Now().Sub(c.start)
	if err != nil {
		c.Error = err.Error()
		c.errorReported = true
	}
	c.Unlock()
}

//StartServiceCall create a ServiceCall timer. ServiceCall timers are used to capture
//the times of calls to services/backends made within a Contributor timing. Note that
//while an error can be denoted when ending the call, it is not assumed an error
//in a call to a service causes the end to end timing to fail.
func (c *Contributor) StartServiceCall(name string, endpoint string) *ServiceCall {
	svcCall := &ServiceCall{
		start:    time.Now(),
		Name:     name,
		Endpoint: endpoint,
	}

	c.Lock()
	c.ServiceCalls = append(c.ServiceCalls, svcCall)
	c.Unlock()

	return svcCall
}

//End stops the clock for a ServiceCall
func (sc *ServiceCall) End(err error) {
	sc.Lock()
	sc.Duration = time.Now().Sub(sc.start)
	if err != nil {
		sc.Error = err.Error()
		sc.errorReported = true
	}
	sc.Unlock()
}

func makeTxnId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
