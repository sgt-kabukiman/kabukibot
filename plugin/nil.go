package plugin

type nilWorker struct{}

func (nw *nilWorker) Enable() {
	// do nothing
}

func (nw *nilWorker) Disable() {
	// do nothing
}

func (nw *nilWorker) Part() {
	nw.Disable()
}

func (nw *nilWorker) Shutdown() {
	nw.Disable()
}

func (nw *nilWorker) Permissions() []string {
	return []string{}
}
