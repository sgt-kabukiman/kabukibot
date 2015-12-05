package plugin

type nilWorker struct{}

func (nw *nilWorker) Enable()   {}
func (nw *nilWorker) Disable()  {}
func (nw *nilWorker) Part()     {}
func (nw *nilWorker) Shutdown() {}

func (nw *nilWorker) Permissions() []string {
	return []string{}
}
