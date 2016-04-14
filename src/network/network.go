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
const BROADCAST_IP string = "129.241.187.255"

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
				if recMsg.MessageTo == localIP || recMsg.MessageTo == BROADCAST_IP {
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

func Slave(elevator ElevatorInfo, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo, addToRequestsChannel chan ButtonInfo) {
	slaveIP := findLocalIPAddress()
	StatusChannel <- "IP of slave is: " + slaveIP

	receivedUdpMessageChannel := make(chan Message, 1)
	go ReceiveUdpMessage(receivedUdpMessageChannel, slaveIP)

	messageFromMasterChannel := make(chan Message, 1)
	masterIsDeadChannel := make(chan bool, 1)
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

		case messageFromMaster := <-messageFromMasterChannel:
			masterIP = messageFromMaster.MessageFrom
			StatusChannel <- "		Updated masterIP = " + masterIP

			msgToMaster := Message{FromMaster: false, AcknowledgeMessage: true, NewOrder: false, ElevatorInfoUpdate: false, MessageTo: masterIP, MessageFrom: slaveIP, ElevatorInfo: elevator, ButtonInfo: ButtonInfo{0, 0, 0}}
			SendUdpMessage(msgToMaster)

			StatusChannel <- "		AcknowledgeMessage sent to master"

			if messageFromMaster.NewOrder {
				StatusChannel <- "		NewOrder"
				addToRequestsChannel <- messageFromMaster.ButtonInfo
			}

		case <-masterIsDeadChannel:
			go Master(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel)
			return
		}
	}
}

func messageFromMaster(receivedUdpMessageChannel chan Message, messageFromMasterChannel chan Message, masterIsDeadChannel chan bool) {
	for {
		select {
		case messageFromMaster := <-receivedUdpMessageChannel:
			if messageFromMaster.FromMaster {
				messageFromMasterChannel <- messageFromMaster
			}

		case <-After(200 * Millisecond):
			masterIsDeadChannel <- true
			return
		}
	}
}

func Master(elevator ElevatorInfo, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo, addToRequestsChannel chan ButtonInfo) {
	masterIP := findLocalIPAddress()
	StatusChannel <- "IP of master is: " + masterIP

	receivedUdpMessageChannel := make(chan Message, 1)
	slaveIsAliveChannel := make(chan Message, 1)
	findBestElevatorForTheJobChannel := make(chan ButtonInfo, 1)
	slavesAliveMapIsChangedChannel := make(chan map[string]ElevatorInfo)
	thisIsTheBestElevatorChannel := make(chan string)
	masterElevatorInfoChannel := make(chan ElevatorInfo, 1)

	go ReceiveUdpMessage(receivedUdpMessageChannel, masterIP)
	go slaveTracker(slaveIsAliveChannel, masterIP, elevator, slavesAliveMapIsChangedChannel)
	go orderHandler.BestElevatorForTheJob(findBestElevatorForTheJobChannel, slavesAliveMapIsChangedChannel, thisIsTheBestElevatorChannel, masterElevatorInfoChannel, masterIP)

	statusMessageToSlave := Message{true, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}}
	SendUdpMessage(statusMessageToSlave)

	for {
		//StatusChannel <- "In Master: "

		select {
		case updatedMasterElevatorInfo := <-updateElevatorInfoChannel:
			masterElevatorInfoChannel <- updatedMasterElevatorInfo

		case receivedMessage := <-receivedUdpMessageChannel:
			if receivedMessage.MessageFrom == masterIP {
				break
			}

			//StatusChannel <- "Received message from IP: " + receivedMessage.MessageFrom
			slaveIsAliveChannel <- receivedMessage

			if receivedMessage.NewOrder {
				StatusChannel <- "		NewOrder"
				externalOrderChannel <- receivedMessage.ButtonInfo
			} else if receivedMessage.ElevatorInfoUpdate {
				StatusChannel <- "		ElevatorInfoUpdate, CurrentFloor = " + strconv.Itoa(receivedMessage.ElevatorInfo.CurrentFloor)
			} else if receivedMessage.AcknowledgeMessage {
				//StatusChannel <- "		AcknowledgeMessage from Slave received"
			}

		case newExternalOrder := <-externalOrderChannel:
			StatusChannel <- "	externalOrderChannel"

			findBestElevatorForTheJobChannel <- newExternalOrder
			bestElevatorIP := <-thisIsTheBestElevatorChannel

			StatusChannel <- "		bestElevatorIP = " + bestElevatorIP
			if bestElevatorIP == masterIP {
				StatusChannel <- "			bestElevatorIP == masterIP"
				addToRequestsChannel <- newExternalOrder
				break
			}
			StatusChannel <- "			bestElevatorIP = Slave"
			msgToSlave := Message{true, false, true, false, masterIP, bestElevatorIP, elevator, newExternalOrder}
			SendUdpMessage(msgToSlave)

		case <-After(100 * Millisecond):

			//statusMessageToSlave := Message{true, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}}
			SendUdpMessage(statusMessageToSlave)
			//sendAliveMessageToSlavesChannel <- true

		}

	}
}

func slaveTracker(slaveIsAliveChannel chan Message, masterIP string, elevator ElevatorInfo, slavesAliveMapIsChangedChannel chan map[string]ElevatorInfo) {

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
				slavesAliveMapIsChangedChannel <- slavesAliveMap
				StatusChannel <- "New slave was added in list of slaves"

				go slaveWatchdog(aliveMessage.MessageFrom, slaveWatchdogChannelsMap[aliveMessage.MessageFrom], terminateSlaveChannel)
			} else {
				slaveWatchdogChannelsMap[aliveMessage.MessageFrom] <- true
			}

		case slaveToBeTerminatedIP := <-terminateSlaveChannel:
			StatusChannel <- "slaveTracker terminates slave with IP: " + slaveToBeTerminatedIP
			delete(slavesAliveMap, slaveToBeTerminatedIP)
			delete(slaveWatchdogChannelsMap, slaveToBeTerminatedIP)
			slavesAliveMapIsChangedChannel <- slavesAliveMap

			/*
				case <-sendAliveMessageToSlavesChannel:

					for slaveIP, _ := range slavesAliveMap {
						statusMessageToSlave := Message{true, false, false, false, masterIP, slaveIP, elevator, ButtonInfo{0, 0, 0}}
						SendUdpMessage(statusMessageToSlave)
					}
			*/
		}

	}
}

func slaveWatchdog(slaveIP string, slaveIsAliveChannel chan bool, terminateSlaveChannel chan string) {
	StatusChannel <- "slaveWatchdog thread created"
	for {
		select {
		case <-slaveIsAliveChannel:
			break

		case <-After(500 * Millisecond):
			terminateSlaveChannel <- slaveIP
			StatusChannel <- "slaveWatchdog timed out, slaveWatchdog thread terminated"
			return
		}
	}
}

func findLocalIPAddress() string {
	ifaces, _ := net.Interfaces()
	var ip net.IP
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				if len(ip.String()) == 15 {
					return ip.String()
				}
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
		}
	}
	return ip.String()
}
