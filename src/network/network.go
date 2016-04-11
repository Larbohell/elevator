package network

import "net"
import "elevator_type"
import "time"
import "encoding/json"

//func recieveUdpMessage(master bool, responseChannel chan source.Message, terminate chan bool, terminated chan int){
func ReceiveUdpMessage(responseChannel chan elevator_type.Message){
	
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
	var recMsg elevator_type.Message
	for {
		_ = recievesock.SetReadDeadline(time.Now().Add(50*time.Millisecond))
		//source.ErrorChannel <- err
		select{
			/*
			case <- terminate:
				err := recievesock.Close()
				source.ErrorChannel <- err
				terminated <- 1
				return
			*/
			default:
				mlen , _, _ := recievesock.ReadFromUDP(buffer)
				if(mlen > 0){
					_ = json.Unmarshal(buffer[:mlen], &recMsg)
					//source.ErrorChannel <- err
					responseChannel <- recMsg
				}
		}
	}
}

func SendUdpMessage(msg elevator_type.Message){
	baddr,err := net.ResolveUDPAddr("udp", "129.241.187.255:27000")


	if(msg.FromMaster){
		baddr,err = net.ResolveUDPAddr("udp", "129.241.187.255:26969")
	}
	sendSock, err := net.DialUDP("udp", nil ,baddr) 
	buf, err:= json.Marshal(msg)
	_,err = sendSock.Write(buf)

	if( err != nil){
		//source.ErrorChannel <- err
	}	
}