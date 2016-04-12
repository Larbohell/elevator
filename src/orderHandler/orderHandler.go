package orderHandler

import "elevator_type"

func OrderHandler(addOrderChannel chan elevator_type.ButtonInfo, removeOrderChannel chan elevator_type.ButtonInfo, addToRequestsChannel chan elevator_type.ButtonInfo) {

	for {
		select {
		case addOrder := <-addOrderChannel:
			if addOrder.Button == elevator_type.BUTTON_INSIDE_COMMAND {
				addToRequestsChannel <- addOrder
			} else {
				// Inform master, master handles order and adds to one of the elevator's queues
			}
		}
	}
}
