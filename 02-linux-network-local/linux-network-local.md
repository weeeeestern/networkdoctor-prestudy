# [Week01] 네트워크 기초 조사 내용 - 동욱

## 1. 오늘 조사한 주제
- Linux network troubleshooting

## 2. 조사 내용 요약
- To troubleshoot Linux network latency, use ping and mtr for RTT/packet loss diagnostics, traceroute to identify slow hops, dig for DNS speed checks, and tcpdump for packet analysis. Common fixes include analyzing network paths, checking for saturation, and optimizing TCP buffers or disabling CPU power-saving C-states. 

Key Tools and Techniques

Ping (ping <IP>): Basic check for round-trip time (RTT) and packet loss.
MTR (mtr <IP>): Combines traceroute and ping to identify exactly which node is causing high latency.
Traceroute (traceroute <IP>): Maps the network path to detect slow, intermediate routers.
Tcpdump (tcpdump -i eth0): Captures traffic to identify network bottlenecks.
Dig (dig <domain>): Tests DNS resolution speed. 

Common Causes and Fixes

High Traffic/Congestion: Check for saturated links.
Pathing Issues: Use mtr to detect routing changes.
CPU Power Saving: Disable C-states to improve network latency. - 얘는 빼도 될듯함 (최저 성능이 아닌 최고 성능을 위한 것)
Configuration Errors: Check for mismatched MTU sizes or bad cabling.
Buffer Tuning: Tune net.core.rmem_max and net.core.wmem_max. 

System Monitoring

Use top or htop to check if high server load is causing slow application response.
Utilize {Link: Site24x7 tools https://www.site24x7.com/learn/linux/network-performance-troubleshooting.html} for in-depth analysis. 

## 3. 참고 자료 (Links)
- [wafaicloud](https://wafaicloud.com/blog/troubleshooting-slow-ssh-connections-in-linux/)
- [serverfault] (https://serverfault.com/questions/445077/how-to-troubleshoot-latency-between-2-linux-hosts)
- [RedHatDocs] (https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/network_troubleshooting_and_performance_tuning/improving-the-network-latency)
- [ManageEngineSite] (https://www.site24x7.com/learn/linux/network-performance-troubleshooting.html)
- [Codilime] (https://codilime.com/blog/linux-network-troubleshooting/)
- [interserver] (https://www.interserver.net/tips/kb/how-to-troubleshoot-and-fix-slow-ssh-connections-in-linux/)
- [netally] (https://www.netally.com/network-performance/what-is-network-latency-and-how-to-reduce-it/)
- [reddit] (https://www.reddit.com/r/networking/comments/jncxsu/how_do_you_go_about_troubleshooting_latency/)
- [scaler] (https://www.scaler.com/topics/linux-for-networking/)

- [리눅스네트워크커맨드시트] (https://www.geeksforgeeks.org/linux-unix/linux-network-commands-cheat-sheet/)


## 다음 조사
## Linux network troubleshooting

- Linux network troubleshooting involves checking physical connectivity, IP configuration (ip a), DNS settings, and routing (ip route) using command-line tools. Key diagnostics include testing connectivity with ping, tracing paths with mtr or traceroute, and inspecting traffic with ss, netstat, or wireshark. Logs in /var/log/syslog and dmesg are crucial for identifying errors. 

Essential Troubleshooting Steps
Physical/Link Layer: Check cable connections and link lights. Use ip link show to verify the interface is "UP".
IP Configuration: Verify the interface has an IP address using ip addr show.
Local Connectivity: Ping the localhost (ping 127.0.0.1) and the local default gateway to isolate issues to the machine or the network.
DNS Resolution: Test if domain names resolve using dig or nslookup.
Routing: View the routing table with ip route to check default gateway configuration.
Service/Port Status: Check if a service is listening on a port using ss -tulpn or netstat -tulpn. 

Common Command-Line Tools
ip (iproute2): Modern tool for interface (ip a), routing (ip r), and link (ip l) management.
ping: Tests connectivity to a host.
mtr / traceroute: Diagnoses network path, latency, and packet loss.
ss / netstat: Examines socket connections.
nmcli: Manages NetworkManager on RHEL/CentOS/Fedora.
tcpdump / wireshark: Captures and analyzes network packets. 

Common Issues & Solutions
Firewall Blocking: Check iptables -L or ufw status for restricted traffic.
Interface Down: Bring up an interface with sudo ip link set <interface> up.
MTU Issues: If packets drop, check and adjust the Maximum Transmission Unit (MTU) with ip link set dev <interface> mtu <size>.
Restart Services: Restart network services using sudo systemctl restart NetworkManager or networking

- https://unix.stackexchange.com/questions/50098/linux-network-troubleshooting-and-debugging#:~:text=Here%20are%20some%20tips%20for%20troubleshooting%20and,%60openssl%60%20*%20%60wireshark%60%20*%20%60iftop%60%20*%20%60iptstate%60
- https://www.reddit.com/r/linuxquestions/comments/1f7ob2m/how_do_i_troubleshoot_network_problems_on_linux/
- https://www.facebook.com/groups/2249498416/posts/10164525101148417/
- https://cycle.io/learn/troubleshooting-linux-networking
- https://www.redhat.com/en/blog/beginners-guide-network-troubleshooting-linux#:~:text=This%20article%20covers%20basic%20network%20troubleshooting%20using,address%60%20command%20*%20**Layer%204:%20Transport%20layer**
- https://www.linuxjournal.com/content/troubleshooting-network-problems#:~:text=If%20you%20can't%20reach%20a%20network%20resource%2C,resource%2C%20you%20probably%20have%20a%20networking%20problem
- https://help.automox.com/hc/en-us/articles/31580375076628-Troubleshooting-Connection-Issues-on-Linux
- https://www.freecodecamp.org/news/how-troubleshoot-network-on-linux/

