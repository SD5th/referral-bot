package receivers

import (
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
)

func NewUpdateReceiver(core *core.Core) (interfaces.UpdateReceiver, error) {

	var updateReceiver interfaces.UpdateReceiver
	var err error

	config := core.GetConfig()
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	switch config.Receiver.Type {
	case "poller":
		updateReceiver, err = NewPoller(core)
		if err != nil {
			return nil, err
		}
	case "webhook":
		updateReceiver, err = NewWebhook(core)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown UpdateReceiverType")
	}

	return updateReceiver, nil
}

type UpdateReceiverConfig struct {
	AllowedUpdates []string
}
