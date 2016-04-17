package network

import "net"
import . "elevator_type"

import . "time"
import "encoding/json"

import "strconv"

import "orderHandler"
import . "statusHandler"

import "math/rand"

//import "fmt"

//import . "errorHandler"

const PORT string = ":24200"
const PORT_FOR_BACKUP_PROCESS string = ":11114"

const BROADCAST_IP string = "129.241.187.255"

//func recieveUdpMessage(master bool, responseChannel chan source.Message, terminate chan bool, terminated chan int){
func ReceiveUdpMessage(receivedUdpMessageChannel chan Message, localIP string, terminateThreadChannel chan bool, threadIsTerminatedChannel chan bool) {
	//StatusChannel <- "In ReceiveUdpMessage function, localIP: " + localIP

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
	receivesock, err := net.ListenUDP("udp", raddr)

	if err != nil {
		StatusChannel <- "ERROR IN LISTENUDP"
		StatusChannel <- err.Error() // PRINT TO ERROR
	}
	var recMsg Message
	for {
		_ = receivesock.SetReadDeadline(Now().Add(50 * Millisecond))
		select {

		case terminate := <-terminateThreadChannel:
			terminateThreadChannel <- terminate
			receivesock.Close()
			StatusChannel <- "Slave " + localIP + ": receiving socket closed and receivedUdpMessage thread shut down"
			threadIsTerminatedChannel <- true
			return

		default:
			mlen, _, _ := receivesock.ReadFromUDP(buffer)
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

func ReceiveUdpMessageOnOwnIP(receivedUdpMessageChannel chan Message, terminateThreadChannel chan bool, threadIsTerminatedChannel chan bool) {
	localIP := findLocalIPAddress()
	//StatusChannel <- "In ReceiveUdpMessageOnOwnIP function, localIP: " + localIP

	buffer := make([]byte, 4098)
	//raddr, err := net.ResolveUDPAddr("udp", ":26969")	// Master port
	raddr, err := net.ResolveUDPAddr("udp", PORT_FOR_BACKUP_PROCESS)
	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
	}

	/*
		source.ErrorChannel <- err
		if(master){
			raddr, err = net.ResolveUDPAddr("udp", ":27000")
			source.ErrorChannel <- err
		}
	*/
	receivesock, err := net.ListenUDP("udp", raddr)
	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
	}

	var recMsg Message
	for {
		_ = receivesock.SetReadDeadline(Now().Add(200 * Millisecond))
		select {

		case terminate := <-terminateThreadChannel:
			terminateThreadChannel <- terminate
			receivesock.Close()
			StatusChannel <- "Slave " + localIP + ": receiving socket closed and receivedUdpMessage thread shut down"
			threadIsTerminatedChannel <- true
			return

		default:
			mlen, _, err := receivesock.ReadFromUDP(buffer)
			if err != nil {
				StatusChannel <- "------------------- HERE"
				StatusChannel <- err.Error() // PRINT TO ERROR
			}
			if mlen > 0 {
				_ = json.Unmarshal(buffer[:mlen], &recMsg)
				//source.ErrorChannel <- err
				//StatusChannel <- "	Received message"
				if recMsg.MessageFrom == localIP {
					receivedUdpMessageChannel <- recMsg
				}
			}
		}
	}
}

func SendUdpMessage(msg Message) {

	//SEND MESSAGE TO SLAVES/MASTER
	baddr, err := net.ResolveUDPAddr("udp", msg.MessageTo+PORT)

	sendSock, err := net.DialUDP("udp", nil, baddr)
	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
		//sendSock.Close()
		return
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
	}
	_, err = sendSock.Write(buf)
	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
	}
	sendSock.Close()

	//SEND ALIVEMESSAGE TO BACKUP PROCESS THAT MONITORS PRIMARY PROCESS
	baddrForBackupProcess, err := net.ResolveUDPAddr("udp", msg.MessageTo+PORT_FOR_BACKUP_PROCESS)

	sendSockForBackupProcess, err := net.DialUDP("udp", nil, baddrForBackupProcess)

	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
		//sendSock.Close()
		return
	}

	bufForBackupProcess, err := json.Marshal(msg)
	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
	}
	_, err = sendSockForBackupProcess.Write(bufForBackupProcess)
	if err != nil {
		StatusChannel <- err.Error() // PRINT TO ERROR
	}
	sendSockForBackupProcess.Close()
}

func Slave(elevator ElevatorInfo, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo, addToRequestsChannel chan ButtonInfo, uncompletedExternalOrders [N_FLOORS][N_BUTTONS - 1]string, orderCompletedByThisElevatorChannel chan ButtonInfo, uncompletedExternalOrdersMatrixChangedChannel chan [N_FLOORS][N_BUTTONS - 1]string) {
	slaveIP := findLocalIPAddress()
	StatusChannel <- "Slave is alive with IP: " + slaveIP

	messageFromMasterChannel := make(chan Message, 1)
	masterIsDeadChannel := make(chan bool, 1)
	receivedUdpMessageChannel := make(chan Message, 1)
	terminateUdpReceiveThreadChannel := make(chan bool, 1)
	udpReceiveThreadIsTerminatedChannel := make(chan bool, 1)

	//StatusChannel <- "Seed is: " + slaveIP[12:15]
	seedNumber, _ := strconv.Atoi(slaveIP[12:15])
	randomSeed := rand.NewSource(int64(seedNumber))
	rand.New(randomSeed)

	go ReceiveUdpMessage(receivedUdpMessageChannel, slaveIP, terminateUdpReceiveThreadChannel, udpReceiveThreadIsTerminatedChannel)
	go messageFromMaster(receivedUdpMessageChannel, messageFromMasterChannel, masterIsDeadChannel)

	//var MasterIP string
	//MasterIP := "129.241.187.156"
	masterIP := "0"

	for {
		//StatusChannel <- "In Slave: "

		select {
		case newExternalOrder := <-externalOrderChannel:
			StatusChannel <- "	In Slave: externalOrderChannel"

			msgToMaster := Message{false, false, true, false, false, slaveIP, masterIP, elevator, newExternalOrder, uncompletedExternalOrders}
			SendUdpMessage(msgToMaster)

		case elevator = <-updateElevatorInfoChannel:
			StatusChannel <- "	In Slave: updateElevatorInfoChannel"
			if masterIP == "0" {
				break
			}
			msgToMaster := Message{false, false, false, false, true, slaveIP, masterIP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
			SendUdpMessage(msgToMaster)

		case messageFromMaster := <-messageFromMasterChannel:
			masterIP = messageFromMaster.MessageFrom
			//StatusChannel <- "		Updated masterIP = " + masterIP

			msgToMaster := Message{false, true, false, false, false, slaveIP, masterIP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
			SendUdpMessage(msgToMaster)
			//StatusChannel <- "Message sent from slave to master, to IP: " + masterIP
			//StatusChannel <- "		AcknowledgeMessage sent to master"

			if messageFromMaster.NewOrder {
				//StatusChannel <- "		NewOrder from master"
				if messageFromMaster.MessageTo == slaveIP {
					StatusChannel <- "		Order is given to slave by master"
					addToRequestsChannel <- messageFromMaster.ButtonInfo

				} else if messageFromMaster.MessageTo == BROADCAST_IP {
					StatusChannel <- "		new updated uncompletedExternalOrderList received from master"
					uncompletedExternalOrders = messageFromMaster.UncompletedExternalOrders
					StatusChannel <- "		uncompletedExternalOrders updated"
					uncompletedExternalOrdersMatrixChangedChannel <- uncompletedExternalOrders
					StatusChannel <- "		Slave should update lights based on unCompletedExternalOrders list"

				}

			}

		case button := <-orderCompletedByThisElevatorChannel:
			StatusChannel <- "In Slave: OrderCompletedByTHisElevatorChannel"
			msgToMaster := Message{false, false, false, true, false, slaveIP, masterIP, elevator, button, uncompletedExternalOrders}
			SendUdpMessage(msgToMaster)

		case <-masterIsDeadChannel:
			StatusChannel <- "In slave: Master is dead"
			terminateUdpReceiveThreadChannel <- true
			<-udpReceiveThreadIsTerminatedChannel

			go Master(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel, uncompletedExternalOrders, orderCompletedByThisElevatorChannel, uncompletedExternalOrdersMatrixChangedChannel)

			return
		}
	}
}

func messageFromMaster(receivedUdpMessageChannel chan Message, messageFromMasterChannel chan Message, masterIsDeadChannel chan bool) {
	for {
		select {
		case messageFromMaster := <-receivedUdpMessageChannel:
			if messageFromMaster.FromMaster {
				//StatusChannel <- "Got messageFromMaster"
				messageFromMasterChannel <- messageFromMaster
			}

		case <-After(Duration(200 * Millisecond)):
			//StatusChannel <- "Did not get messageFromMaster, timeout after " + strconv.Itoa(200+rand.Intn(20)) + "ms, shutting down slave"
			masterIsDeadChannel <- true
			return
		}
	}
}

func Master(elevator ElevatorInfo, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo, addToRequestsChannel chan ButtonInfo, uncompletedExternalOrders [N_FLOORS][N_BUTTONS - 1]string, orderCompletedByThisElevatorChannel chan ButtonInfo, uncompletedExternalOrdersMatrixChangedChannel chan [N_FLOORS][N_BUTTONS - 1]string) {
	counter := 0

	masterIP := findLocalIPAddress()
	StatusChannel <- "Master is alive with IP: " + masterIP

	receivedUdpMessageChannel := make(chan Message, 1)
	slaveIsAliveChannel := make(chan Message, 1)
	externalOrderFromDeadSlaveChannel := make(chan ButtonInfo, 1)

	slaveIsAliveIPChannel := make(chan string, 1) //Merge slaveIsAliveIPCHannel and slaveIsAliveCHannel

	orderCompletedChannel := make(chan ButtonInfo, 1)

	slavesAliveMapIsChangedChannel := make(chan map[string]ElevatorInfo) //////////////////////////////////////////////UNBUFFERED?
	findBestElevatorForTheJobChannel := make(chan ButtonInfo, 1)
	thisIsTheBestElevatorChannel := make(chan string) //////////////////////////////////////////////UNBUFFERED?
	masterElevatorInfoChannel := make(chan ElevatorInfo, 1)
	terminateThreadChannel := make(chan bool, 1)
	threadIsTerminatedChannel := make(chan bool, 1)

	go ReceiveUdpMessage(receivedUdpMessageChannel, masterIP, terminateThreadChannel, threadIsTerminatedChannel)
	go slaveTracker(slaveIsAliveChannel, masterIP, elevator, slavesAliveMapIsChangedChannel, terminateThreadChannel, threadIsTerminatedChannel)
	go orderHandler.BestElevatorForTheJob(findBestElevatorForTheJobChannel, slavesAliveMapIsChangedChannel, thisIsTheBestElevatorChannel, masterElevatorInfoChannel, masterIP, terminateThreadChannel, threadIsTerminatedChannel)
	numberOfThreads := 3

	statusMessageToSlave := Message{true, false, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
	SendUdpMessage(statusMessageToSlave)

	// You may have been a slave in an earlier life, so:
	// Iterate over unserved external order matrix, put all unserved orders in externalOrderChannel (buffered with N_FlOORS * 2 - 2 elements)
	//Should not spawn orderWatchdogs for on itself

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			if uncompletedExternalOrders[floor][btn] != "" && uncompletedExternalOrders[floor][btn] != masterIP {
				var button ButtonInfo
				button.Button = Button(btn)
				button.Floor = floor
				button.Value = 1

				//Obsolete, put on newExternalOrder
				uncompletedExternalOrders[floor][btn] = ""
				externalOrderChannel <- button
				//go orderWatchdog(uncompletedExternalOrders[floor][btn], button, slaveIsAliveIPChannel, externalOrderFromDeadSlaveChannel, orderCompletedChannel)

			}
		}
	}

	for {
		//StatusChannel <- "In Master: "

		select {
		case elevator = <-updateElevatorInfoChannel:
			StatusChannel <- "updateElevatorInfoChannel"

			masterElevatorInfoChannel <- elevator

			statusMessageToSlave = Message{true, false, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
			SendUdpMessage(statusMessageToSlave)

		case receivedMessage := <-receivedUdpMessageChannel:
			if receivedMessage.MessageFrom == masterIP {
				break
			}
			//StatusChannel <- "a"

			//StatusChannel <- "Received message from IP: " + receivedMessage.MessageFrom
			if !receivedMessage.FromMaster {
				slaveIsAliveChannel <- receivedMessage
				select {
				case <-slaveIsAliveIPChannel:
				default:
				}
				slaveIsAliveIPChannel <- receivedMessage.MessageFrom

			} else {
				StatusChannel <- "Master got message from another master, fights to the death for reign over the elevator-kingdom"
				myThreeLastNumbersOfIP, _ := strconv.Atoi(masterIP[12:15])
				receivedThreeLastNumbersOfIP, _ := strconv.Atoi(receivedMessage.MessageFrom[12:15])

				if myThreeLastNumbersOfIP > receivedThreeLastNumbersOfIP {
					terminateThreadChannel <- true

					for i := 0; i < numberOfThreads; i++ {
						<-threadIsTerminatedChannel
					}
					<-terminateThreadChannel // emptying channel
					go Slave(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel, uncompletedExternalOrders, orderCompletedByThisElevatorChannel, uncompletedExternalOrdersMatrixChangedChannel)
					return
				}
			}
			//StatusChannel <- "b"

			if receivedMessage.NewOrder {

				StatusChannel <- "		NewOrder"
				externalOrderChannel <- receivedMessage.ButtonInfo

			} else if receivedMessage.ElevatorInfoUpdate {
				StatusChannel <- "		ElevatorInfoUpdate, CurrentFloor = " + strconv.Itoa(receivedMessage.ElevatorInfo.CurrentFloor)

			} else if receivedMessage.AcknowledgeMessage {
				//StatusChannel <- "		AcknowledgeMessage from Slave received"

			} else if receivedMessage.OrderCompleted {
				StatusChannel <- "			Ordercompleted in Master"

				// Empty orderCompletedChannel if not empty
				select {
				case <-orderCompletedChannel:
					break
				default:
				}

				orderCompletedChannel <- receivedMessage.ButtonInfo
				uncompletedExternalOrders[receivedMessage.ButtonInfo.Floor][int(receivedMessage.ButtonInfo.Button)] = ""
				msgToSlave := Message{true, false, false, true, false, masterIP, BROADCAST_IP, elevator, receivedMessage.ButtonInfo, uncompletedExternalOrders}
				SendUdpMessage(msgToSlave)

				StatusChannel <- "Updated uncompletedExternalOrders sent to slave from Master"
				uncompletedExternalOrdersMatrixChangedChannel <- uncompletedExternalOrders //Update own lights
				StatusChannel <- "Master should update lights based on unCompletedExternalOrders list"

				// Clears button lights in all slaves (and itself), and remove order from unserved external orders matrix and kill orderWatchdog
			}
			//StatusChannel <- "c"
			statusMessageToSlave = Message{true, false, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
			SendUdpMessage(statusMessageToSlave)
			//StatusChannel <- "d"

		case newExternalOrder := <-externalOrderChannel:
			StatusChannel <- "In Master: case newExternalOrder"
			counter++
			//StatusChannel <- "3"

			if uncompletedExternalOrders[newExternalOrder.Floor][int(newExternalOrder.Button)] != "" {
				statusMessageToSlave = Message{true, false, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
				StatusChannel <- "external Order already in list, does not add"
				SendUdpMessage(statusMessageToSlave)

				break
			}

			findBestElevatorForTheJobChannel <- newExternalOrder
			bestElevatorIP := <-thisIsTheBestElevatorChannel
			StatusChannel <- strconv.Itoa(counter) + ": bestElevatorIP: " + bestElevatorIP

			// Create orderWatchdog thread, which monitors that order, times out if slave not alive for some time, and terminates if order is served

			if bestElevatorIP == masterIP {
				StatusChannel <- "			bestElevatorIP == masterIP"
				addToRequestsChannel <- newExternalOrder
				StatusChannel <- strconv.Itoa(counter) + ": bestElevator = Master, added to addToRequestsChannel"
			} else {
				//Order best slave to excecute order
				StatusChannel <- "			bestElevatorIP = Slave"

				go orderWatchdog(bestElevatorIP, newExternalOrder, slaveIsAliveIPChannel, externalOrderFromDeadSlaveChannel, orderCompletedChannel)

				msgToSlave := Message{true, false, true, false, false, masterIP, bestElevatorIP, elevator, newExternalOrder, uncompletedExternalOrders}
				SendUdpMessage(msgToSlave)

			}

			//Update uncompletedExternalOrders in all slaves

			uncompletedExternalOrders[newExternalOrder.Floor][newExternalOrder.Button] = bestElevatorIP
			uncompletedExternalOrdersMatrixChangedChannel <- uncompletedExternalOrders //Update own lights

			msgToSlaves := Message{true, false, true, false, false, masterIP, BROADCAST_IP, elevator, newExternalOrder, uncompletedExternalOrders}
			SendUdpMessage(msgToSlaves)
			StatusChannel <- strconv.Itoa(counter) + ": END OF EXTERNALORDERCHANNEL"

		case button := <-orderCompletedByThisElevatorChannel:
			// NEVER GETS HERE
			StatusChannel <- "orderCompletedByThisElevatorChannel"
			//StatusChannel <- "4"
			uncompletedExternalOrders[button.Floor][int(button.Button)] = ""
			msgToSlave := Message{true, false, false, true, false, masterIP, BROADCAST_IP, elevator, button, uncompletedExternalOrders}
			SendUdpMessage(msgToSlave)

			uncompletedExternalOrdersMatrixChangedChannel <- uncompletedExternalOrders //Update own lights

		case oldExternalOrder := <-externalOrderFromDeadSlaveChannel:
			StatusChannel <- "externalOrderFromDeadSlaveChannel--------------------------------------------------------------"
			//StatusChannel <- "5"
			uncompletedExternalOrders[oldExternalOrder.Floor][int(oldExternalOrder.Button)] = ""
			externalOrderChannel <- oldExternalOrder

			statusMessageToSlave = Message{true, false, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
			SendUdpMessage(statusMessageToSlave)

		case <-After(100 * Millisecond):

			//statusMessageToSlave := Message{true, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}}
			//StatusChannel <- "Broadcasting"
			statusMessageToSlave = Message{true, false, false, false, false, masterIP, BROADCAST_IP, elevator, ButtonInfo{0, 0, 0}, uncompletedExternalOrders}
			SendUdpMessage(statusMessageToSlave)
			//sendAliveMessageToSlavesChannel <- true

		}

	}
}

func slaveTracker(slaveIsAliveChannel chan Message, masterIP string, elevator ElevatorInfo, slavesAliveMapIsChangedChannel chan map[string]ElevatorInfo, terminateThreadChannel chan bool, threadIsTerminatedChannel chan bool) {

	slavesAliveMap := make(map[string]ElevatorInfo)
	slaveWatchdogChannelsMap := make(map[string]chan bool)

	terminateSlaveChannel := make(chan string)
	StatusChannel <- "Number of slaves at the beginning: " + strconv.Itoa(len(slavesAliveMap))

	for {
		select {
		case aliveMessage := <-slaveIsAliveChannel:
			_, IPexsistsInSlavesAliveMap := slavesAliveMap[aliveMessage.MessageFrom]
			//StatusChannel <- "slaveIsAliveChannel action from slave with IP " + aliveMessage.MessageFrom

			slavesAliveMap[aliveMessage.MessageFrom] = aliveMessage.ElevatorInfo
			slavesAliveMapIsChangedChannel <- slavesAliveMap

			if !IPexsistsInSlavesAliveMap {

				newWatchdogChannel := make(chan bool, 1)
				slaveWatchdogChannelsMap[aliveMessage.MessageFrom] = newWatchdogChannel

				StatusChannel <- "Slave with IP " + aliveMessage.MessageFrom + " was added to list of slaves, number of slaves is now: " + strconv.Itoa(len(slavesAliveMap))

				go slaveWatchdog(aliveMessage.MessageFrom, slaveWatchdogChannelsMap[aliveMessage.MessageFrom], terminateSlaveChannel)
			} else {
				slaveWatchdogChannelsMap[aliveMessage.MessageFrom] <- true
			}

		case slaveToBeTerminatedIP := <-terminateSlaveChannel:
			delete(slavesAliveMap, slaveToBeTerminatedIP)
			delete(slaveWatchdogChannelsMap, slaveToBeTerminatedIP)
			StatusChannel <- "slaveTracker terminates slave with IP: " + slaveToBeTerminatedIP + ", number of slaves: " + strconv.Itoa(len(slavesAliveMap))

			slavesAliveMapIsChangedChannel <- slavesAliveMap

		case terminate := <-terminateThreadChannel:
			terminateThreadChannel <- terminate
			StatusChannel <- "slaveTracker is terminated"
			threadIsTerminatedChannel <- true
			return

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

		case <-After(200 * Millisecond):
			terminateSlaveChannel <- slaveIP
			StatusChannel <- "slaveWatchdog timed out, slaveWatchdog thread terminated"
			return
			/*
				case terminate <- terminateThreadChannel:
					terminateThreadChannel <- terminate
					StatusChannel <- "slaveWatchdog is terminated"
					threadIsTerminatedChannel <- true
					return
			*/
		}
	}
}

func orderWatchdog(slaveIP string, button ButtonInfo, slaveIsAliveIPChannel chan string, externalOrderFromDeadSlaveChannel chan ButtonInfo, orderCompletedChannel chan ButtonInfo) {
	StatusChannel <- "orderWatchdog thread created"
	for {
		select {
		case receivedSlaveIP := <-slaveIsAliveIPChannel:
			// HAS TO BE PUT BACK, OR DONE SOME OTHER WAY (MAYBE BY SENDING BOTH IP AND BUTTON), BECAUSE MORE THAN ONE ORDERWATCHDOG CAN BE INTERESTED IN WHETHER ONE
			// SLAVE IS ALIVE

			//////////////////////////////////// IMPORTANTE //////////////////////////////////
			if receivedSlaveIP == slaveIP {
				//StatusChannel <- "Received ALIVE from slave"
				break
			} else {
				slaveIsAliveIPChannel <- receivedSlaveIP
			}

		case receivedButton := <-orderCompletedChannel:
			if receivedButton == button {
				StatusChannel <- "Order completed, orderWatchdog terminates, IP " + slaveIP + "with order button " + strconv.Itoa(int(receivedButton.Button)) + " in floor " + strconv.Itoa(receivedButton.Floor)
				return
			}
			orderCompletedChannel <- receivedButton

		case <-After(300 * Millisecond):
			externalOrderFromDeadSlaveChannel <- button
			StatusChannel <- "orderWatchdog timed out, orderWatchdog thread terminated"
			return
			/*
				case terminate <- terminateThreadChannel:
					terminateThreadChannel <- terminate
					StatusChannel <- "slaveWatchdog is terminated"
					threadIsTerminatedChannel <- true
					return
			*/
		}
	}
}

func findLocalIPAddress() string {
	ifaces, _ := net.Interfaces()
	var ip net.IP
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
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
		}
	}
	//Send to ErrorChannel, could not find IP
	return ""
}
