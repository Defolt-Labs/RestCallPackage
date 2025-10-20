package network

// HTTP Methods
const (
	methodGET     = "GET"
	methodPOST    = "POST"
	methodPUT     = "PUT"
	methodDELETE  = "DELETE"
	methodPATCH   = "PATCH"
	methodHEAD    = "HEAD"
	methodOPTIONS = "OPTIONS"
)

// Logging constants
const (
	DottedSeparator   = "................................................."
	LogFormat         = "[%s] %-20s %s"
	LogFormatInt      = "[%s] %-20s %d"
	LogRequestDesc    = "Request Description:"
	LogHttpMethod     = "HTTP Method:"
	LogDestEndpoint   = "Destination Endpoint:"
	LogPayload        = "Payload:"
	LogHeaders        = "Headers:"
	LogResponseDesc   = "Response Description:"
	LogResponseStatus = "Response Status:"
	LogResponse       = "Response:"
	LogNullValue      = "null"
)
