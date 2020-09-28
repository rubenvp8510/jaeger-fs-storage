package filesystem

import (
	"flag"
	"github.com/spf13/viper"
)

const (
	pathFlag                  = "fsstorage.path"
	ephemeralFlag             = "fsstorage.ephemeral"
	numberReadWorkersFlag     = "fsstorage.numReadWorkers"
	numberWriteWorkersFlag    = "fsstorage.numWriteWorkers"
	writeBufferSizeFlag       = "fsstorage.WriteBufferSize"
	defaultNumberReadWorkers  = 100
	defaultNumberWriteWorkers = 100
	defaultNumberWriteBuffer  = 50
)

type Options struct {
	Configuration Config
}

func AddFlags(flagSet *flag.FlagSet) {
	flagSet.String(pathFlag, ".", "Path to store traces information.")
	flagSet.Bool(ephemeralFlag, true, "Ephemeral store data on a temp file")
	flagSet.Int(numberReadWorkersFlag, defaultNumberReadWorkers, "Number of read workers")
	flagSet.Int(numberWriteWorkersFlag, defaultNumberWriteWorkers, "Number of write workers")
	flagSet.Int(writeBufferSizeFlag, defaultNumberWriteBuffer, "Size of write buffer")
}

// InitFromViper initializes the options struct with values from Viper
func (opt *Options) InitFromViper(v *viper.Viper) {
	opt.Configuration.Ephemeral = v.GetBool(ephemeralFlag)
	opt.Configuration.DataDir = v.GetString(pathFlag)
	opt.Configuration.NumberWriteWorkers = v.GetInt(numberWriteWorkersFlag)
	opt.Configuration.NumberReadWorkers = v.GetInt(numberReadWorkersFlag)
	opt.Configuration.WriteBufferSize = v.GetInt(writeBufferSizeFlag)
}
