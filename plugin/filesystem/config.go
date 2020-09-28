package filesystem

type Config struct {
	DataDir            string
	Ephemeral          bool
	NumberReadWorkers  int
	NumberWriteWorkers int
	WriteBufferSize    int
}

