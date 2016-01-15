/*
Package timer implements a simple timing utility that can be used to capture
end to end timings, plus any subtimings of interests.
 */
package timer

import (
	"encoding/json"
	"time"
	"sync"
)

//ServiceCall is used to capture a service call made in the context of a Contributor timing.
type ServiceCall struct {
	Name  string
	Time  time.Duration
	Error string
	start time.Time
}

//Contributor is used to capture sub-timings of note that contribute to the end to end time.
type Contributor struct {
	sync.Mutex
	Name         string
	Time         time.Duration
	Error        string
	start        time.Time
	ServiceCalls []*ServiceCall
}

//EndToEndTimer is used to capture an end to end timing. Subtimings can be added to
//an end to end time using StartContributor.
type EndToEndTimer struct {
	sync.Mutex
	Name         string
	Time         time.Duration
	Contributors []*Contributor
	ErrorFree    bool
	Error        string
	start        time.Time
}

//NewEndToEndTimer creates a new EndToEndTimer. Note that the clock starts when an
//EndToEndTimer is created.
func NewEndToEndTimer(name string) *EndToEndTimer {
	return &EndToEndTimer{
		Name:  name,
		start: time.Now(),
	}
}

//Stop stops the clock for an EndToEndTimer. If an error occurred in the timing path,
//it should be noted by passing an error object to Stop, otherwise pass nil. The JSON
//representation of the timing data will reflect if an error occurred during the timing.
func (t *EndToEndTimer) Stop(err error) {
	t.Time = time.Now().Sub(t.start)
	if err != nil {
		t.Error = err.Error()
	}
	t.ErrorFree = len(t.ContributorErrors()) == 0 && t.Error == ""
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

//ContributorErrors produces a slice of all errors that have been
//reported by the  contributor subtimings associated with an
//EndToEndTimer
func (t *EndToEndTimer) ContributorErrors() []string {
	var errs []string
	for _, c := range t.Contributors {
		if c.Error != "" {
			errs = append(errs, c.Error)
		}
	}
	return errs
}

//ToJSONString produces a JSON string representation of an EndToEndTimer, including all of
//its contributors.
func (t *EndToEndTimer) ToJSONString() string {
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
	c.Time = time.Now().Sub(c.start)
	if err != nil {
		c.Error = err.Error()
	}
}

//StartServiceCall create a ServiceCall timer. ServiceCall timers are used to capture
//the times of calls to services/backends made within a Contributor timing. Note that
//while an error can be denoted when ending the call, it is not assumed an error
//in a call to a service causes the end to end timing to fail.
func (c *Contributor) StartServiceCall(name string) *ServiceCall {
	svcCall := &ServiceCall{
		start: time.Now(),
		Name:  name,
	}

	c.Lock()
	c.ServiceCalls = append(c.ServiceCalls, svcCall)
	c.Unlock()

	return svcCall
}

//End stops the clock for a ServiceCall
func (sc *ServiceCall) End(err error) {
	sc.Time = time.Now().Sub(sc.start)
	if err != nil {
		sc.Error = err.Error()
	}
}
