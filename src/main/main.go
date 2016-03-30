package main
import "network"
import "fmt"
import "strconv"
import "flag"
import "os/exec"

const aliveMessage = "I'm alive!(Lars&Bendik)"
const aliveMessage_length = 23
const timer_reset_value = 100
var primary bool

// Problem: Backup doesn't pick up where the primary left, but starts counting from 0.
// This comes from a problem with converting between strings and ints, that results in
// last_number_received always being interpreted as 0


func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
    }
}

func main() {
	
	flag.BoolVar(&primary, "primary", true, "lol")
	flag.Parse()

	starting_point := 0;

	for{
		// Primary
		if (primary == true){
			spawn_backup();
			count_and_broadcast(starting_point)
		
		// Backup
		} else {
			starting_point = listen()
		}
	}
}

func listen() int {
	timer := timer_reset_value
	last_number_received := 0

	for {
		message := network.Udp_listen(":20011");
		
		if (message[:aliveMessage_length] != aliveMessage){
			timer--
			//fmt.Println("Timer Decrement")
		} else if (message[:aliveMessage_length] == aliveMessage) {
			timer = timer_reset_value
			last_number_received, _ = strconv.Atoi(message[aliveMessage_length:])
			//CheckError(err)
			fmt.Println(message[aliveMessage_length:])
			//fmt.Println("Last num: " + strconv.Itoa(last_number_received))
			//fmt.Println(aliveMessage)
		} 
		if (timer == 0){
			fmt.Println("Timer == 0")
			primary = true

			fmt.Println("Last num: " + strconv.Itoa(last_number_received))
			return last_number_received
		}
	}
	
}

func count_and_broadcast(starting_point int){
	i := starting_point
	for{
		if (i%100==0) {
			network.Udp_broadcast(aliveMessage + strconv.Itoa(i))
			fmt.Println("Sent: " + strconv.Itoa(i));
			//fmt.Println(aliveMessage + strconv.Itoa(i))
		}
	
		i++
	}
}

func spawn_backup(){
	run_process_command := exec.Command("gnome-terminal", "-x", "./main", "-primary=0")
	err := run_process_command.Start()
	CheckError(err)
}