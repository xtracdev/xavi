package service

import (
	"crypto/rand"
	"encoding/hex"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"time"
)

const (
	nominal          = iota
	connectPoolError = iota
	backendCallError = iota
)

const (
	perflogTag = "perflog"
)

type ServiceTimer struct {
	state            int
	txnId            string
	uri              string
	soapAction       string
	serviceStartTime time.Time
	serviceEndTime   time.Time
	backendStartTime time.Time
	backendEndTime   time.Time
	status           int
	err              error
}

//Generate a random txn id
func generateTxnID() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		log.Warn("random generator is generating an error: ", err.Error())
		return "xxx"
	}

	return hex.EncodeToString(buf)
}

func NewServiceTimer(req *http.Request) *ServiceTimer {
	st := new(ServiceTimer)
	st.txnId = generateTxnID()
	st.state = nominal
	st.uri = req.RequestURI
	st.soapAction = req.Header.Get("SOAPAction")
	st.serviceStartTime = time.Now()
	return st
}

func (st *ServiceTimer) ConnectFail(err error) {
	if err == nil {
		panic("ServiceTimer ConnectFail called with nil error")
	}
	st.err = err
	st.state = connectPoolError
}

func (st *ServiceTimer) BackendCallStart() {
	st.backendStartTime = time.Now()
}

func (st *ServiceTimer) BackendCallEnd(err error) {
	st.backendEndTime = time.Now()
	if err != nil {
		st.err = err
		st.state = backendCallError
	}
}

func (st *ServiceTimer) EndService(status int) {
	st.serviceEndTime = time.Now()

	switch st.state {
	case nominal:
		log.WithFields(log.Fields{
			"msgtype":     perflogTag,
			"txnid":       st.txnId,
			"uri":         st.uri,
			"soapAction":  st.soapAction,
			"serviceTime": st.serviceEndTime.Sub(st.serviceStartTime),
			"backendTime": st.backendEndTime.Sub(st.backendStartTime),
			"status":      status,
		}).Info()

	case connectPoolError:
		log.WithFields(log.Fields{
			"msgtype":     perflogTag,
			"txnid":       st.txnId,
			"uri":         st.uri,
			"soapAction":  st.soapAction,
			"serviceTime": st.serviceEndTime.Sub(st.serviceStartTime),
			"status":      status,
			"error":       st.err.Error(),
		}).Info()

	case backendCallError:
		log.WithFields(log.Fields{
			"msgtype":     perflogTag,
			"txnid":       st.txnId,
			"uri":         st.uri,
			"soapAction":  st.soapAction,
			"serviceTime": st.serviceEndTime.Sub(st.serviceStartTime),
			"backendTime": st.backendEndTime.Sub(st.backendStartTime),
			"status":      status,
			"error":       st.err.Error(),
		}).Info()
	}

}
