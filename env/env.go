package env

//Environment variable names used to pick up configuration are defined here.
const (
	KVStoreURL     = "XAVI_KVSTORE_URL"
	LoggingOpts    = "XAVI_LOGGING_OPTS"
	StatsdEndpoint = "XAVI_STATSD_ADDRESS"
	LoggingLevel   = "XAVI_LOGGING_LEVEL"
)

//Valid LoggingOpts - note that all can be specified in the env var, comma
//separated.
const (
	Tcplog = "TCPLOG"
)

//Environment variable for specifying a non-default TCP log address.
//Note the expected form of this address is host:port
const (
	TcplogAddress = "XAVI_TCPLOG_ADDRESS"
)

//Address to listen to http pprof requests on, for example localhost:6060
const PProfEndpoint = "XAVI_PPROF_ENDPOINT"
