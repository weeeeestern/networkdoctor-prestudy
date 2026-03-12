# [Week01] 네트워크 기초 조사 내용 - 동욱

## 1. 오늘 조사한 주제
- Linux 네트워크 트러블슈팅(문제 해결)

## 2. 조사 내용 요약
Linux 네트워크 지연 문제를 진단 및 해결하는 방법에는 여러 가지가 있다.
ping과 mtr을 사용하여 RTT/패킷 손실을 진단하고, traceroute를 통해 지연이 발생하는 홉(hop)을 식별하며, dig를 사용해 DNS 속도를 확인하고, tcpdump로 패킷을 분석한다. 
일반적인 해결 방법으로는 네트워크 경로 분석, 대역폭 포화(saturation) 상태 확인, TCP 버퍼 최적화, CPU 절전 C-state 비활성화 등이 있다.

주요 도구 및 기술

Ping : 왕복 지연 시간(RTT) 및 패킷 손실에 대한 기본적인 확인.
MTR (MyTraceRoute) : traceroute와 ping을 결합하여 어느 노드에서 높은 지연이 발생하는지 정확히 식별.
Traceroute: 네트워크 경로를 매핑하여 속도가 느린 중간 라우터를 감지.
Tcpdump (tcpdump -i eth0): 트래픽을 캡처하여 네트워크 병목 현상을 파악.
Dig (domain) : DNS 확인(분석) 속도를 테스트.

주요 원인 및 해결 방법

높은 트래픽/혼잡: 포화된 링크(대역폭)가 있는지 확인.
경로 문제: mtr을 사용하여 라우팅 변경 사항을 감지.
CPU 절전 모드: 네트워크 지연 시간을 개선하기 위해 C-state를 비활성화. - cpu 절전 모드를 없애는 것
구성 오류: MTU 크기 불일치 또는 불량 케이블 확인.
버퍼 튜닝: net.core.rmem_max 및 net.core.wmem_max 값을 조정.

시스템 모니터링

top 또는 htop을 사용하여 서버 부하가 높아 애플리케이션 응답이 느려지는지 확인. -> 향후 cilium 통합 예정

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