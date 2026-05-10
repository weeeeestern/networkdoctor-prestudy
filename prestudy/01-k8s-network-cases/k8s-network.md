- Pod 생성
- CNI가 IPAM과 함께 Pod 에 IP 부여
- kube-proxy가 Service와 실제 Pod 연결해주는 규칙 관리자
- iptables, IPVS, eBPF가 실제 패킷을 보고 DNAT 및 패킷 전달 수행

---

# CNI

Container Network Interface

쿠버네티스가 네트워크를 직접 구현하지 않고, 표준 인터페이스만 정의하여 실제 구현은 외부 플러그인 (Calico, Flannel, Cilium) 에 맡긴다.

- kublet/runtime이 **새 Pod를 만들으려 함**
- CNI 플러그인 호출
    - 네트워크 네임스페이스 준비
    - veth pair 생성
    - IPAM과 함께 Pod IP 할당
    - 라우팅 설정
    - overlay/BGP/eBPF 연결

## network namespace

Pod는 자신만의 네트워크 ns를 갖는다.

따라서 자신만의 eth0, routing table, loopback 을 갖는다.

→ 서로 다른 Pod들이 같은 포트를 사용할 수 있고, 네트워크 설정이 분리된 채, 독립된 프로세스처럼 동작할 수 있게 된다.

## veth pair

한쪽 끝은 Pod, 다른 쪽 끝은 host → 한 쌍의 가상 이더넷 인터페이스

Pod 안에서는 host와 연결돼있는 veth pair의 한 쪽이 **eth0**으로 보인다.

## bridge

이 veth들이 보통 bridge와 연결된다. cni0 같은 거.. 가상 스위치

같은 노드 안 Pod 내부 통신 : Pod A → veth(eth0) → bridge(cni0) → veth(eth0) → Pod B

# 다른 노드의 Pod 간 통신

## Native Routing

Pod IP CIDR을 기본 네트워크가 직접 라우팅 할 수 있는 방식

encapsulation 없고 단순해서 성능 좋음

but 물리or 클라우드 네트워크가 Pod CIDR 라우팅을 이해해야 함

## Overlay

기본 네트워크가 Pod CIDR 를 몰라도 Pod 간 통신을 가능하게 해주는 방식

### VXLAN

- L2 over UDP 느낌
- 네트워크 독립성

### IPIP

- IP packet을 또 IP packet으로 감싸는 방식
- Calico에서 자주 언급됨

**원래 Pod-to-Pod 패킷을 Node-to-Node 패킷 안에 넣어 운반**
: 기본 네트워크가 Pod IP를 모르니, 패킷을 Node IP 패킷 안에 감싸서 보낸다.

기본 네트워크에 덜 의존할 수 있고 어디서든 쉽게 동작할 수 있다.

but encapsulation, decapsulation 비용 발생 → CPU 오버헤드 발생 가능

## BGP

각 노드들이, 자신이 어떤 Pod의 대역을 가지고 있는지 광고 전파하는 방식

encapsulation 필요 없다, 대규모 환경에 잘 맞다

but 네트워크 운영 난이도 높아진다..

# Service

## Kube-proxy

API Server를 watch 하고 있다가, Service 내부 설정의 변경을 감지하면

노드 커널 안의 iptables or IPVS 기술을 활용하여 네트워크 규칙을 생성/수정/관리.

ingress 트래픽의  Dest IP (Service IP)를 Pod IP로 DNAT하여 목적지까지 전달하는건 커널 안의 기술들과 CNI!!!

- 요즘에는 kube-proxy를 아예 빼고, Cilium 같은 CNI 플러그인이 eBPF라는 커널 기술로 해당 라우팅을 직접 처리

### iptables

iptables 기반 kube-proxy는

Service IP로 온 패킷을 실제 Pod로 DNAT 하는 규칙 체인을 만든다.

- 패킷이 서비스의 ClusterIP로 들어옴
- 각각, 해당 규칙 체인들에서 서비스를 매칭하고 실제 Pod IP로 DNAT 하고…

** ingress - DNAT,    egress - SNAT // Service는 DNAT

O(n) → 대규모 환경에서 안 조음

### IPVS

리눅스 커널에 내장된 L4 로드밸런서 -  해시 테이블 기반의 규칙

O(1) 빠름. 빠른데도,

IPVS 기반 kube-proxy는 최신 k8s 버전에서는 deprecated 됐다 이유는?

- IPVS는 로드밸런싱만 수행하기 때문에, SNAT이나 NodePort 트래픽을 처리하지 못 한다. 따라서 IPVS 모드여도 여전히 iptables나 nftables을 부가적으로 사용했어야 했다.
- IPVS는 리눅스의 Netfilter를 다 거치고 나서야 IPVS구나 하고 처리가 된다. 하지만 Cilium은 네트워크 드라이버의 바로 위 XDP or 최하단 tc에서만 패킷을 낚아챈다.
- 파드 새로 생겼을 때도 IPVS는 연관된 모든 것들을 다 갱신해야하는데, eBPF는 공유 메모리인 Map을 사용하기에 해당 부분만 갱신하면 된다.

# conntrack

connection tracking

원래 네트워크의 기본 통신 IP 패킷은 Stateless 입니다. 하지만 conntrack은 리눅스 커널의 네트워크 스택 안에서 동작하며, 오가는 모든 패킷을 연결하여 메모리에 기록합니다.

TCP같은 Stateful한 패킷뿐아니라, UDP 같은 Stateless한 프로토콜의 패킷에도 가상의 상태를 부여하여 추적합니다.

- NEW : ex. TCP의 SYN 등
- ESTABLISHED : 양방향 패킷 오고간게 확인된 상태
- RELATED : 기존의 ESTABLISHED과 연관이 있는 새로운 연결 생성 패킷
- INVALID : 어떤 연결에도 속하지 않은 비정상 패킷

ingress 패킷이 노드의 커널을 통과할 때 DNAT이 수행될 때 → conntrack 기록

egress 할때 다시 Pod IP 에서→ 서비스 IP로 빠져나가야 클라이언트가 해당 패킷을 수락한다. → SNAT 작업 (Reverse DNAT)을 하려면 기록이 필요하다 그것이 conntrack

# Service

외부 → LB/NodePort → kube-proxy/IPVS/eBPF → Pod

## ClusterIP

클러스터 내부에서만 쓰는 **가상** 고정 IP

## NodePort

ClusterIP에서 +a: 모든 노드의 특정 포트를 외부에 개방

노드 IP:특정 포트 로 요청이 들어오면, nodeport 서비스를 통해 파드로 전달

## LoadBalancer

NodePort에서 +a: AWS, GCP 같은 클라우드의 로드밸런서까지 붙인 서비스

외부 트래픽은 먼저 로드밸런서로 들어오고, 이후 노드와 파드로 전달됨

## 서비스 ExternalTrafficPolicy : Cluster

외부(클라이언트)에서 어떤 노드로 들어와도 다른 노드의 Pod로 항상 넘길 수 있다.

- 클라가 파드에 접속하길 원해서 1번 노드에 접속함 but 1번 노드에는 파드가 없고 2번노드에 파드가 있음
- 1번 노드는 2번 노드로 트래픽을 넘길 수 있다.

하지만 그렇게 넘기면 1번 노드가 응답도 받을 수 있어야 하기 때문에, Source를 클라이언트에서 1번 노드로 SNAT 하고 보낸다

⇒ 원본 클라이언트의 IP가 보존되지 않을 수 있다.

## 서비스 ExternalTrafficPolicy : Local

원본 클라이언트 IP를 보존하자 그리고 불필요한 Network Hop을 줄이자

나라는 노드에 들어오는 트래픽을 절대 다른 노드로 전달하지마!!!!

그래서 단점이 드롭될 수 있다. 그래서 보통은 Local 설정된 서비스의 앞단에 로드밸런서를 둬서, 넘겨야하는 노드가 있으면 애초애 그 노드로는 보내지 말도록 설정해둔다고 한다.

# nftables

리눅스 커널의 네트워크 스택의 패킷이 지나다니는 곳들마다 netfilter라는 hook이 존재하고, iptables는 그 hook 마다 rule을 적어둔다

iptables은 단순히 텍스트만 나열하는 형태로 규칙을 작성해뒀다면, nftables은 규칙들을 컴파일하여 커널 내부의 작은 가상머신에서 실행되는 바이트코드 형태로 변환한다.

얘도 IPVS처럼 해시테이블을 사용하기에 O(1)이지만, IPVS는 라우팅 역할만 한다!

nftables는 방화벽, NAT, 패킷분류 역할을 수행한다.

# eBPF

얘는 iptables 처럼 리눅스 내장 기능도 아니고, netfilter 마다 규칙을 적지 않는다.

사용자가 eBPF 코드를 직접 작성하여 커널 내부에 삽입 후 직접 실행시키는 기술이다!!

아예 kube-proxy를 사용하지 않음

### XDP

NIC 드라이버에 있음

패킷이 NIC에 도착하자마자 리눅스의 네트워크 스택을 거치기도 전에 패킷을 가로채서 즉시 폐기 or 목적지 변경 가능 → 엄청나게 빠르다.

근데 ingress 패킷만 처리할 수 있음

### tc

얘는 그래도 리눅스 네트워크 스택 안의 훅중 하나인데,

얘는 ingress랑 egress 모두에 ebpf 를 걸 수 있다!

그래서 egress나 복잡한 규칙이 필요한 패킷들은 tc에서 처리함

---

# 헷갈려 : CNI 플러그인과 **iptables/nftables/eBPF의 관계**

### control plane

쿠버네티스 API Server에 A랑 B Pod를 연결하시오, C Service를 만드시오 명령 입력됨

CNI 플러그인은, 네트워크 패킷을 실제로 만지는 게 아니라, 쿠버네티스의 명령을 리눅스가 이해할 수 있게 번역해서 커널에 설정값을 주입하는 것

### data plane

리눅스의 Kernel Space에 존재하는 네트워크 처리 엔진들 == iptables, nftables, eBPF

NIC에 들어온 패킷을 검사하고 목적지로 보내고 방화벽 역할도 하고…

해당 커널 엔진들은 독립적인 판단 능력이 없다. CNI 플러그인이 규칙을 넣어주면 그걸 실행하는 것이다.