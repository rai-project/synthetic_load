package synthetic_load

type Runner interface {
	// the input is a sequence of bytes and an
	// on completion function
	Run(TraceEntry, []byte, func()) error
}
