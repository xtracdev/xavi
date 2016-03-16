package timer

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPostitiveDuration(t *testing.T) {
	at := NewEndToEndTimer("foo")
	at.Stop(nil)
	if at.Duration == 0 {
		t.Fail()
	}

}

func TestContributors(t *testing.T) {
	at := NewEndToEndTimer("foo")
	c1 := at.StartContributor("c1")
	c2 := at.StartContributor("c2")
	c2.End(nil)
	c1.End(nil)
	at.Stop(nil)

	if at.Error != "" {
		t.Fail()
	}

	if c1.Duration <= 0 || c2.Duration <= 0 {
		t.Fail()
	}

	if at.ErrorFree == false {
		t.Fail()
	}
}

func TestIfContributorErrorsThenTimerErrors(t *testing.T) {
	at := NewEndToEndTimer("foo")
	c1 := at.StartContributor("c1")
	c2 := at.StartContributor("c2")
	c2.End(errors.New("oh whoops"))
	c1.End(nil)
	at.Stop(nil)

	if at.Error != "" {
		t.Fail()
	}

	if len(at.ContributorErrors()) != 1 {
		t.Fail()
	}

	if at.ErrorFree == true {
		t.Fail()
	}

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

	c3.End(nil)

	c2.End(nil)
	c1.End(nil)
	at.Stop(nil)

	if at.Error != "" {
		t.Fail()
	}

	if c1.Duration <= 0 || c2.Duration <= 0 || c3.Duration <= 0 {
		t.Fail()
	}

	if at.ErrorFree == false {
		t.Fail()
	}

	if len(c3.ServiceCalls) != 2 {
		t.Fail()
	}

	println(at.ToJSONString())
}

func TestEndToEndTimerRaceCondition(t *testing.T) {
	eet := NewEndToEndTimer("racy")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		eet.StartContributor("fst")
		time.AfterFunc(1e9, func() { eet.Stop(nil) })
	}()

	wg.Add(1)
	time.AfterFunc(2e9, func() {
		defer wg.Done()
		s := eet.ToJSONString()
		t.Logf("%s\n", s)
	})
	wg.Wait()
}

func TestEndToEndTimerContributorRaceCondition(t *testing.T) {
	eet := NewEndToEndTimer("racy")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c := eet.StartContributor("fst")
		sc := c.StartServiceCall("foo call", "/foo/service")
		wg.Add(3)
		time.AfterFunc(1e9, func() { defer wg.Done(); sc.End(nil) })
		time.AfterFunc(2e9, func() { defer wg.Done(); c.End(fmt.Errorf("Error num %v", 1)) })
		time.AfterFunc(3e9, func() { defer wg.Done(); eet.Stop(nil) })
	}()

	wg.Add(1)
	time.AfterFunc(2e9, func() {
		defer wg.Done()
		s := eet.ToJSONString()
		t.Logf("%s\n", s)
	})

	errs := eet.ContributorErrors()
	t.Logf("%v\n", errs)

	wg.Wait()
}

func TestNewEndToEndTimerRaceCondition(t *testing.T) {
	var eet *EndToEndTimer
	var wg sync.WaitGroup

	eet = NewEndToEndTimer("foo")

	wg.Add(1)
	go func() {
		defer wg.Done()
		eet.StartContributor("fst")
	}()

	wg.Wait()
	t.Logf("%s\n", eet.ToJSONString())

}
