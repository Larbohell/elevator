package network

import "net"
import . "elevator_type"

import . "time"
import "encoding/json"
import "strconv"
import "orderHandler"
import . "statusHandler"

//import . "errorHandler"

const PORT string = ":24541"

//func recieveUdpMessage(master bool, responseChannel chan source.Message, terminate chan bool, terminated chan int){
func ReceiveUdpMessage(receivedUdpMessageChannel chan Message, localIP string) {
	StatusChannel <- "In ReceiveUdpMessage function"

	buffer := make([]byte, 4098)
	//raddr, err := net.ResolveUDPAddr("udp", ":26969")	// Master port
	raddr, _ := net.ResolveUDPAddr("udp", PORT)
	/*
		source.ErrorChannel <- err
		if(master){
			raddr, err = net.ResolveUDPAddr("udp", ":27000")
			source.ErrorChannel <- err
		}
	*/
	recievesock, _ := net.ListenUDP("udp", raddr)
	//source.ErrorChannel <- err
	var recMsg Message
	for {
		_ = recievesock.SetReadDeadline(Now().Add(50 * Millisecond))
		//source.ErrorChannel <- err
		select {
		/*
			case <- terminate:
				err := recievesock.Close()
				source.ErrorChannel <- err
				terminated <- 1
				return
		*/
		default:
			mlen, _, _ := recievesock.ReadFromUDP(buffer)
			if mlen > 0 {
				_ = json.Unmarshal(buffer[:mlen], &recMsg)
				//source.ErrorChannel <- err
				//StatusChannel <- "	Received message"
				if recMsg.MessageTo == localIP {
					receivedUdpMessageChannel <- recMsg
				}
			}
		}
	}
}

func SendUdpMessage(msg Message) {

	baddr, err := net.ResolveUDPAddr("udp", msg.MessageTo+PORT)

	/*
		if msg.FromMaster {
			baddr, err = net.ResolveUDPAddr("udp", "129.241.187.255:26969")
		}
	*/

	sendSock, err := net.DialUDP("udp", nil, baddr)
	buf, err := json.Marshal(msg)
	_, err = sendSock.Write(buf)

	if err != nil {
		//ErrorChannel <- "SendUdpMessage: Error."
	}
}

func Slave(elevator ElevatorInfo, slaveIP string, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo, addToRequestsChannel chan ButtonInfo) {
	receivedUdpMessageChannel := make(chan Message, 1)
	go ReceiveUdpMessage(receivedUdpMessageChannel, slaveIP)

	//var MasterIP string
	//MasterIP := "129.241.187.156"
	masterIP := "0"

	for {
		StatusChannel <- "In Slave: "

		select {
		case newExternalOrder := <-externalOrderChannel:
			StatusChannel <- "	externalOrderChannel"

			msgToMaster := Message{false, false, true, false, slaveIP, masterIP, elevator, newExternalOrder}
			SendUdpMessage(msgToMaster)

		case elevator = <-updateElevatorInfoChannel:
			StatusChannel <- "	updateElevatorInfoChannel"
			if masterIP == "0" {
				break
			}
			msgToMaster := Message{FromMaster: false, AcknowledgeMessage: false, NewOrder: false, ElevatorInfoUpdate: true, MessageTo: masterIP, MessageFrom: slaveIP, ElevatorInfo: elevator, ButtonInfo: ButtonInfo{0, 0, 0}}
			SendUdpMessage(msgToMaster)

		case messageFromMaster := <-receivedUdpMessageChannel:
			StatusChannel <- "	receivedUdpMessageChannel"
			if messageFromMaster.FromMaster {

				masterIP = messageFromMaster.MessageFrom
				StatusChannel <- "		Updated masterIP = " + masterIP

				msgToMaster := Message{FromMaster: false, AcknowledgeMessage: true, NewOrder: false, ElevatorInfoUpdate: false, MessageTo: masterIP, MessageFrom: slaveIP, ElevatorInfo: elevator, ButtonInfo: ButtonInfo{0, 0, 0}}
				SendUdpMessage(msgToMaster)

				StatusChannel <- "		AcknowledgeMessage sent to master"

				if messageFromMaster.NewOrder {
					StatusChannel <- "		NewOrder"
					addToRequestsChannel <- messageFromMaster.ButtonInfo
				}
			}
		}
	}
}

func Master(elevator ElevatorInfo, masterIP string, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo, addToRequestsChannel chan ButtonInfo) {
	receivedUdpMessageChannel := make(chan Message, 1)
	slaveIsAliveChannel := make(chan Message, 1)

	go ReceiveUdpMessage(receivedUdpMessageChannel, masterIP)
	go slaveTracker(slaveIsAliveChannel)

	for {
		StatusChannel <- "In Master: "
		select {

		case receivedMessage := <-receivedUdpMessageChannel:
			StatusChannel <- "	receivedUdpMessageChannel"

			slaveIsAliveChannel <- receivedMessage

			if receivedMessage.NewOrder {
				StatusChannel <- "		NewOrder"
				externalOrderChannel <- receivedMessage.ButtonInfo
			} else if receivedMessage.ElevatorInfoUpdate {
				StatusChannel <- "		ElevatorInfoUpdate, CurrentFloor = " + strconv.Itoa(receivedMessage.ElevatorInfo.CurrentFloor)
			} else if receivedMessage.AcknowledgeMessage {
				StatusChannel <- "		AcknowledgeMessage from Slave received"
			}

		case newExternalOrder := <-externalOrderChannel:
			StatusChannel <- "	externalOrderChannel"
			bestElevatorIP := orderHandler.BestElevatorForTheJob(newExternalOrder)
			StatusChannel <- "		bestElevatorIP = " + bestElevatorIP
			if bestElevatorIP == masterIP {
				StatusChannel <- "			bestElevatorIP == masterIP"
				addToRequestsChannel <- newExternalOrder
				break
			}
			StatusChannel <- "			bestElevatorIP = Slave"
			msgToSlave := Message{true, false, true, false, masterIP, bestElevatorIP, elevator, newExternalOrder}
			SendUdpMessage(msgToSlave)

		case <-After(200 * Millisecond):
			slaveIP := "129.241.187.159"

			for i := 0; i < 1; i++ { //Iterate over all available slaves
				statusMessageToSlave := Message{true, false, false, false, masterIP, slaveIP, elevator, ButtonInfo{0, 0, 0}}
				SendUdpMessage(statusMessageToSlave)
			}

		}

	}
}

func slaveTracker(slaveIsAliveChannel chan Message) {

	slavesAliveMap := make(map[string]ElevatorInfo)
	slaveWatchdogChannelsMap := make(map[string]chan bool)

	terminateSlaveChannel := make(chan string)

	for {
		select {
		case aliveMessage := <-slaveIsAliveChannel:
			_, IPexsistsInSlavesAliveMap := slavesAliveMap[aliveMessage.MessageFrom]

			if !IPexsistsInSlavesAliveMap {

				newWatchdogChannel := make(chan bool, 1)

				slaveWatchdogChannelsMap[aliveMessage.MessageFrom] = newWatchdogChannel
				slavesAliveMap[aliveMessage.MessageFrom] = aliveMessage.ElevatorInfo
				StatusChannel <- "New slave was added in list of slaves"

				go slaveWatchdog(aliveMessage.MessageFrom, slaveWatchdogChannelsMap[aliveMessage.MessageFrom], terminateSlaveChannel)
			} else {
				slaveWatchdogChannelsMap[aliveMessage.MessageFrom] <- true
			}

		case slaveToBeTerminatedIP := <-terminateSlaveChannel:
			StatusChannel <- "slaveTracker terminates slave with IP: " + slaveToBeTerminatedIP
			delete(slavesAliveMap, slaveToBeTerminatedIP)
			delete(slaveWatchdogChannelsMap, slaveToBeTerminatedIP)
		}
	}
}

func slaveWatchdog(slaveIP string, slaveIsAliveChannel chan bool, terminateSlaveChannel chan string) {
	StatusChannel <- "slaveWatchdog thread created"
	for {
		select {
		case <-slaveIsAliveChannel:
			break

		case <-After(100 * Millisecond):
			terminateSlaveChannel <- slaveIP
			StatusChannel <- "slaveWatchdog timed out, slaveWatchdog thread terminated"
			return
		}
	}
}
