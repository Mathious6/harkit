package harhandler

type HandlerOption func(*HARHandler)

func WithServerIPAddress() HandlerOption {
	return func(h *HARHandler) {
		h.resolveIPAddress = true
	}
}
