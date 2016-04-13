package network

import "net"
import . "elevator_type"

import . "time"
import "encoding/json"
import "fmt"
import "strconv"
import "orderHandler"

//import . "errorHandler"

//func recieveUdpMessage(master bool, responseChannel chan source.Message, terminate chan bool, terminated chan int){
func ReceiveUdpMessage(receivedUdpMessageChannel chan Message, localIP string) {

	buffer := make([]byte, 4098)
	//raddr, err := net.ResolveUDPAddr("udp", ":26969")	// Master port
	raddr, _ := net.ResolveUDPAddr("udp", ":27000")
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
				if recMsg.MessageTo == localIP {
					receivedUdpMessageChannel <- recMsg
				}
			}
		}
	}
}

func SendUdpMessage(msg Message) {

	baddr, err := net.ResolveUDPAddr("udp", msg.MessageTo+":27000")
	if msg.MessageTo == "allSlaves" {
		baddr, err = net.ResolveUDPAddr("udp", "129.241.187.255:27000")
	}

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
	MasterIP := "0"

	for {
		select {
		case newExternalOrder := <-externalOrderChannel:
			msgToMaster := Message{false, false, true, false, slaveIP, MasterIP, elevator, newExternalOrder}
			SendUdpMessage(msgToMaster)

		case elevator = <-updateElevatorInfoChannel:
			msgToMaster := Message{FromMaster: false, AcknowledgeMessage: false, NewOrder: false, ElevatorInfoUpdate: true, MessageTo: MasterIP, MessageFrom: slaveIP, ElevatorInfo: elevator, ButtonInfo: ButtonInfo{0, 0, 0}}
			SendUdpMessage(msgToMaster)

		case messageFromMaster := <-receivedUdpMessageChannel:
			if messageFromMaster.FromMaster {
				msgToMaster := Message{FromMaster: false, AcknowledgeMessage: true, NewOrder: false, ElevatorInfoUpdate: false, MessageTo: MasterIP, MessageFrom: slaveIP, ElevatorInfo: elevator, ButtonInfo: ButtonInfo{0, 0, 0}}
				SendUdpMessage(msgToMaster)

				MasterIP = messageFromMaster.MessageFrom

				if messageFromMaster.NewOrder {
					addToRequestsChannel <- messageFromMaster.ButtonInfo
				}
			}
		}
	}
}

func Master(elevator ElevatorInfo, masterIP string, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo, addToRequestsChannel chan ButtonInfo) {
	receivedUdpMessageChannel := make(chan Message, 1)
	go ReceiveUdpMessage(receivedUdpMessageChannel, masterIP)

	for {
		select {

		case receivedMessage := <-receivedUdpMessageChannel:
			if receivedMessage.NewOrder {
				externalOrderChannel <- receivedMessage.ButtonInfo
			} else if receivedMessage.ElevatorInfoUpdate {
				fmt.Println("Message received from Slave, Slave's current floor: ", strconv.Itoa(receivedMessage.ElevatorInfo.CurrentFloor))
			}

		case newExternalOrder := <-externalOrderChannel:
			bestElevatorIP := orderHandler.BestElevatorForTheJob(newExternalOrder)

			if bestElevatorIP == masterIP {
				addToRequestsChannel <- newExternalOrder
				break
			}

			msgToSlave := Message{true, false, true, false, masterIP, bestElevatorIP, elevator, newExternalOrder}
			SendUdpMessage(msgToSlave)

		case <-After(10 * Millisecond):
			statusMessageToSlave := Message{true, false, false, false, masterIP, "allSlaves", elevator, ButtonInfo{0, 0, 0}}
			SendUdpMessage(statusMessageToSlave)

		}

	}
}
