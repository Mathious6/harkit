package harhandler

// HandlerOption is a function that can be used to configure a HARHandler
type HandlerOption func(*HARHandler)

// WithServerIPAddress is a function that can be used to configure a HARHandler to resolve the IP address of the server
func WithServerIPAddress() HandlerOption {
	return func(h *HARHandler) {
		h.resolveIPAddress = true
	}
}
