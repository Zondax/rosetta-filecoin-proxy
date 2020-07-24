package tools

import (
	"fmt"
	"time"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("RPC_Lotus")

type LotusRPCWrapper func()

/// WrapWithTimeout executes lotusFunc but return an error after the `timeout`
func WrapWithTimeout(lotusFunc LotusRPCWrapper, timeout time.Duration) error {
	ch := make(chan bool, 1)
	defer func() {
		if ch != nil {
			close(ch)
		}
		ch = nil
	}()

	go func() {
		lotusFunc()
		if ch != nil {
			ch <- true
		}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ch:
		{
			log.Debug("lotus answer received")
			if ch != nil {
				close(ch)
			}
			ch = nil
			return nil
		}
	case <-timer.C:
		{
			log.Error("Lotus RPC call Timed out!")
			if ch != nil {
				close(ch)
			}
			ch = nil
			return fmt.Errorf("Lotus RPC call Timed out!")
		}
	}
}
