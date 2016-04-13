package network

import "net"
import . "elevator_type"

import . "time"
import "encoding/json"
import "fmt"
import "strconv"

//func recieveUdpMessage(master bool, responseChannel chan source.Message, terminate chan bool, terminated chan int){
func ReceiveUdpMessage(receivedUdpMessageChannel chan Message) {

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
				receivedUdpMessageChannel <- recMsg
			}
		}
	}
}

func SendUdpMessage(msg Message) {
	baddr, err := net.ResolveUDPAddr("udp", "129.241.187.255:27000")

	if msg.FromMaster {
		baddr, err = net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	}

	sendSock, err := net.DialUDP("udp", nil, baddr)
	buf, err := json.Marshal(msg)
	_, err = sendSock.Write(buf)

	if err != nil {
		//source.ErrorChannel <- err
	}
}

func Slave(elevator ElevatorInfo, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo) {
	receivedUdpMessageChannel := make(chan Message, 1)
	go ReceiveUdpMessage(receivedUdpMessageChannel)

	for {
		select {
		case newExternalOrder := <-externalOrderChannel:
			msgToMaster := Message{FromMaster: false, NewOrder: true, ElevatorInfoUpdate: false, ElevatorInfo: elevator, ButtonInfo: newExternalOrder}
			SendUdpMessage(msgToMaster)

		case elevator = <-updateElevatorInfoChannel:
			msgToMaster := Message{FromMaster: false, NewOrder: false, ElevatorInfoUpdate: true, ElevatorInfo: elevator, ButtonInfo: ButtonInfo{0, 0, 0}}
			SendUdpMessage(msgToMaster)
		}

	}

}

func Master(elevator ElevatorInfo, externalOrderChannel chan ButtonInfo, updateElevatorInfoChannel chan ElevatorInfo) {
	receivedUdpMessageChannel := make(chan Message, 1)
	go ReceiveUdpMessage(receivedUdpMessageChannel)

	for {
		select {
		case receivedMessage := <-receivedUdpMessageChannel:
			if receivedMessage.FromMaster {
				//Error, msg sent from master to master
			} else if receivedMessage.NewOrder {

			} else if receivedMessage.ElevatorInfoUpdate {
				fmt.Println("Message received from Slave, Slave's current floor: ", strconv.Itoa(receivedMessage.ElevatorInfo.CurrentFloor))
			}
		}

	}
}
