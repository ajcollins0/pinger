package main

import (
	"os/exec"
	"fmt"
    "strconv"
    "strings"
    "os"
    "net"
    "sort"
    "flag"
    "bytes"
    "github.com/brotherpowers/ipsubnet"
)

type machineStatus struct {
	Ip 		string
	Status  string
}

func ping_many(ip string, c chan<- machineStatus){
	_, err := exec.Command("ping", "-c 1", "-W 1", ip).Output()
	if err == nil {
		c <- machineStatus{ip, "online"}
	}else{
		c <- machineStatus{ip, "offline"}
	}
}

func incIp(ip string) string{
    oct := strings.Split(ip, ".")
    o3, _ := strconv.Atoi(oct[3])
	o2, _ := strconv.Atoi(oct[2])
    o1, _ := strconv.Atoi(oct[1])
    o0, _ := strconv.Atoi(oct[0])

    if o3 == 255{
    	o3 = 0
    	if o2 == 255{ 
    		o2 = 0
    		if o1 == 255{
    			o1 = 0
    			if o0 == 255{
    				fmt.Println("Unable to calculate addresses")
    				fmt.Println(o0, o1, o2, o3)
    				os.Exit(1)
    			}else{
    				o0 +=1
    			}
    		}else{
    			o1 +=1
    		}
    	}else{
    		o2 += 1
    	}
    }else{
    	o3 += 1
    }
	return strconv.Itoa(o0) + "." + strconv.Itoa(o1) + "." + strconv.Itoa(o2) + "."+ strconv.Itoa(o3)
}

func calcIps(ip string, mask int)[]string {

	var retrunIps []string
	sub := ipsubnet.SubnetCalculator(ip, mask)

	currentIP := sub.GetIPAddressRange()[0]
	lastIP := sub.GetIPAddressRange()[1]

	retrunIps = append(retrunIps, currentIP)
	for currentIP != lastIP{
		newIp := incIp(currentIP)
		currentIP = newIp
		retrunIps = append(retrunIps, currentIP)
	}
	return retrunIps

}

func splitOffOnline(machines []machineStatus) ([]string, []string){
	var offline []string
	var online []string

	for _, machine := range machines {
		if machine.Status == "offline"{
			offline = append(offline, machine.Ip)
		}else{
			online = append(online, machine.Ip)
		}
	}
	return sortIPArray(online), sortIPArray(offline)
}

func sortIPArray(ips []string)[]string{
	var sorted []string

	realIPs := make([]net.IP, 0, len(ips))

	for _, ip := range ips {
		realIPs = append(realIPs, net.ParseIP(ip))
	}

	sort.Slice(realIPs, func(i, j int) bool {
		return bytes.Compare(realIPs[i], realIPs[j]) < 0
	})

	for _, ip := range realIPs {
		sorted = append(sorted, fmt.Sprintf("%s", ip))
	}
	return sorted
}

func printIps(ips []string){
	for _,ip := range(ips){
		fmt.Println(ip)
	}
}

func getCurrentIP()string{

	var PossibleIPs []string

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Cannot acess Interfaces")
		fmt.Println(err)
		os.Exit(1)
	}
	for _, i := range ifaces {
		flags := fmt.Sprintf("%s",i.Flags)
		// excluding VPN subnets
		if !strings.Contains(flags, "pointtopoint") {
		    addrs, err := i.Addrs()
		    	if err != nil {
					fmt.Println("Cannot acess IP Address from interface:", i)
					fmt.Println(err)
					os.Exit(1)
				}
			    for _, addr := range addrs {
			        var ip net.IP
			        switch v := addr.(type) {
			        case *net.IPNet:
			                ip = v.IP
			        case *net.IPAddr:
			                ip = v.IP
			        }
					// trim out the loop back and ipv6
			        if ip.String() != "127.0.0.1" && !strings.Contains(ip.String(), ":"){
			        	PossibleIPs = append(PossibleIPs, ip.String())
			        }
	   			}
		}
	}
	return PossibleIPs[0]
}

func main() {
	var machines []machineStatus
	c := make(chan machineStatus)

	startIP := flag.String("i", getCurrentIP(), "IP Address to use")
	mask := flag.Int("m", 24, "Address Mask to use")
	showOffline := flag.Bool("offline", false, "Print online assets. Set to false to print offline assets ")

	flag.Parse()

 	ipList := calcIps(*startIP, *mask)

	for _, ip := range ipList{
		go ping_many(ip, c)
	}

	for i := 1; i <= len(ipList); i++ {
		machines = append(machines,<-c)
	}

	on, off := splitOffOnline(machines)

	if *showOffline == true{
		printIps(off)
	}else{
		printIps(on)
	}
}
