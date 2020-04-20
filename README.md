# Ping-CLI-application
A CLI application that sends ICMP echo requests to a given host. I have used golang for this assignment.

I have implemented the basic functionalities like   
1. Send echo requests to a specfic IP Address or a host
2. If the host does not exists, return an error message saying no host exists
3. Display the Packet loss and latency of each message received.
4. Display the total time, total number of packets sent and the total loss once all / desired number of packets are sent

Additional functionalities implemented
1. Distinguish between IPv4 and IPv6 and support both of them
2. Allow user to set TTL via a command line argument
3. Allow the user to set the size of a packet to be sent
4. Quiet ouput suppressing all intermediate messages except the start and finish summaries.
5. Allow the user to send only a `count` number of packets
6. Allow the user to wait for a certain time interval between sending of successive packets
7. Allow the user to specify a timeout before ping exits regardless of how many packets have been sent or received.

# Install necessary packages 
```
go get golang.org/x/net/icmp
go get golang.org/x/net/ipv4
go get golang.org/x/net/ipv6
```  
After this, use the command `go build` and run the program using `sudo ./<dir_name>`   
Here, `dir_name` is the name of the current directory where the go file is located.  
It can also be renamed to ping and can be run with the command `sudo ./ping`


# Usage 

The default hostname is `cloudflare.com` which will be used when no host/IP Address is provided.

## Command line arguments usage instructions

`sudo ./ping -i <interval> -c <count> -w <deadline> -q <quiet output> --ttl <set_ttl> -s <packet_size> <hostname/ipaddress>` where 

```
hostname is a hostname like google.com or facebook.com
ipaddress is any IP Address like 104.17.175.85
interval (i) is an integer (seconds)
count (c) is an integer
deadline (w) is an integer (seconds)
packet_size (s) is an integer which has a max limit of 70 (in bytes)
ttl (ttl) is an integer
quiet (q) is a boolean

```

## Sample Ping Queries

`sudo ./ping google.com` will ping `google.com` indefinitely till Ctrl + C is pressed  
`sudo ./ping 104.17.175.85` will ping the 
`sudo ./ping -c 5` will ping `google.com` and stop after sending 5 packets  
`sudo ./ping -c 5 -i 2 google.com` will ping `google.com` with time interval between successive pings is 2 seconds  
`sudo ./ping -c 5 -q google.com` will ping `google.com` but will not show results of every ping  
`sudo ./ping -c 5 -w 3 google.com` will ping `google.com` only for 3 seconds because the `deadline` is set to 3  

The flags can be combined in any order but the hostname must be specified only after all the flags are specified.     
For example, `sudo ./ping google.com -c 5` will not set the count flag because it is after the hostname.  
