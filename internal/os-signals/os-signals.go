package osSignals

import (
	"os"
	"os/signal"
)

type ChannelWithOsSignals chan os.Signal

func MakeChannelWithInterruptSignal() ChannelWithOsSignals {
	signals := make(ChannelWithOsSignals, 1)
	signal.Notify(signals, os.Interrupt)
	return signals
}
