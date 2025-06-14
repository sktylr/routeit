package routeit

// Http Status codes for responses. https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status
type HttpStatus struct {
	code int
	msg  string
}

var (
	StatusContinue                      = HttpStatus{100, "Continue"}
	StatusSwitchingProtocols            = HttpStatus{101, "Switching Protocols"}
	StatusProcessing                    = HttpStatus{102, "Processing"}
	StatusEarlyHints                    = HttpStatus{103, "Early Hints"}
	StatusOK                            = HttpStatus{200, "OK"}
	StatusCreated                       = HttpStatus{201, "Created"}
	StatusAccepted                      = HttpStatus{202, "Accepted"}
	StatusNonAuthoritativeInformation   = HttpStatus{203, "Non-Authoritative Information"}
	StatusNoContent                     = HttpStatus{204, "No Content"}
	StatusResetContent                  = HttpStatus{205, "Reset Content"}
	StatusPartialContent                = HttpStatus{206, "Partial Content"}
	StatusMultiStatus                   = HttpStatus{207, "Multi-Status"}
	StatusAlreadyReported               = HttpStatus{208, "Already Reported"}
	StatusIMUsed                        = HttpStatus{226, "IM Used"}
	StatusMultipleChoices               = HttpStatus{300, "Multiple Choices"}
	StatusMovedPermanently              = HttpStatus{301, "Moved Permanently"}
	StatusFound                         = HttpStatus{302, "Found"}
	StatusSeeOther                      = HttpStatus{303, "See Other"}
	StatusNotModified                   = HttpStatus{304, "Not Modified"}
	StatusUseProxy                      = HttpStatus{305, "Use Proxy"}
	StatusUnused                        = HttpStatus{306, "Unused"}
	StatusTemporaryRedirect             = HttpStatus{307, "Temporary Redirect"}
	StatusPermanentRedirect             = HttpStatus{308, "Permanent Redirect"}
	StatusBadRequest                    = HttpStatus{400, "Bad Request"}
	StatusUnauthorized                  = HttpStatus{401, "Unauthorized"}
	StatusPaymentRequired               = HttpStatus{402, "Payment Required"}
	StatusForbidden                     = HttpStatus{403, "Forbidden"}
	StatusNotFound                      = HttpStatus{404, "Not Found"}
	StatusMethodNotAllowed              = HttpStatus{405, "Method Not Allowed"}
	StatusNotAcceptable                 = HttpStatus{406, "Not Acceptable"}
	StatusProxyAuthenticationRequired   = HttpStatus{407, "Proxy Authentication Required"}
	StatusRequestTimeout                = HttpStatus{408, "Request Timeout"}
	StatusConflict                      = HttpStatus{409, "Conflict"}
	StatusGone                          = HttpStatus{410, "Gone"}
	StatusLengthRequired                = HttpStatus{411, "Length Required"}
	StatusPreconditionFailed            = HttpStatus{412, "Precondition Failed"}
	StatusContentTooLarge               = HttpStatus{413, "Content Too Large"}
	StatusURITooLong                    = HttpStatus{414, "URI Too Long"}
	StatusUnsupportedMediaType          = HttpStatus{415, "Unsupported Media Type"}
	StatusRangeNotSatisfiable           = HttpStatus{416, "Range Not Satisfiable"}
	StatusExpectationFailed             = HttpStatus{417, "Expectation Failed"}
	StatusImATeapot                     = HttpStatus{418, "I'm a teapot"}
	StatusMisdirectedRequest            = HttpStatus{421, "Misdirected Request"}
	StatusUnprocessableContent          = HttpStatus{422, "Unprocessable Content"}
	StatusLocked                        = HttpStatus{423, "Locked"}
	StatusFailedDependency              = HttpStatus{424, "Failed Dependency"}
	StatusTooEarly                      = HttpStatus{425, "Too Early"}
	StatusUpgradeRequired               = HttpStatus{426, "Upgrade Required"}
	StatusPreconditionRequired          = HttpStatus{428, "Precondition Required"}
	StatusTooManyRequests               = HttpStatus{429, "Too Many Requests"}
	StatusRequestHeaderFieldsTooLarge   = HttpStatus{431, "Request Header Fields Too Large"}
	StatusUnavailableForLegalReasons    = HttpStatus{451, "Unavailable For Legal Reasons"}
	StatusInternalServerError           = HttpStatus{500, "Internal Server Error"}
	StatusNotImplemented                = HttpStatus{501, "Not Implemented"}
	StatusBadGateway                    = HttpStatus{502, "Bad Gateway"}
	StatusServiceUnavailable            = HttpStatus{503, "Service Unavailable"}
	StatusGatewayTimeout                = HttpStatus{504, "Gateway Timeout"}
	StatusHTTPVersionNotSupported       = HttpStatus{505, "HTTP Version Not Supported"}
	StatusVariantAlsoNegotiates         = HttpStatus{506, "Variant Also Negotiates"}
	StatusInsufficientStorage           = HttpStatus{507, "Insufficient Storage"}
	StatusLoopDetected                  = HttpStatus{508, "Loop Detected"}
	StatusNotExtended                   = HttpStatus{510, "Not Extended"}
	StatusNetworkAuthenticationRequired = HttpStatus{511, "Network Authentication Required"}
)
