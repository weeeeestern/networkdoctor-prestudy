# eBPF와 Cilium

## eBPF란 무엇인가

- 운영체제 커널과 같은 특별한 권한이 있는 환경에서 샌드박스 프로그램을 실행시킬 수 있게 해주는 커널 기술
- 샌드박스 프로그램이란 격리된 환경에서 프로그램을 실행시키는 것

운영 체제 커널은 강력한 권한을 가지고 있지만 동시에 안정성과 보안에 중요한 역할을 하여 빠르게 진화하기는 어렵다

하지만 eBPF 프로그램을 통해 커널 코드를 수정하지 않고도 기존의 운영 체제(커널 내부의 특정 지점)에 프로그램을 붙여 실행함으로써 추가적인 기능을 추가할 수 있게 됨

- eBPF = 기술
- eBPF program = 그 기술로 커널에 붙여 실행하는 코드

네트워크, 보안, observablility 등 다양한 영역에서 사용된다.

## eBPF가 왜 필요한가

→ 리눅스 커널 메트릭을 수집하는 프로그램을 실행하기 위해서

**그렇다면 node exporter가 `/proc`, `/sys` 등에서 읽어오는 메트릭으로는 부족한가?**

Node Exporter 방식은 커널 상태 통계만 제공

```go
node_netstat_Tcp_RetransSegs = 10000
```

이를 통해서 TCP 재전송이 많음은 알 수 있으나 어디서 발생했는지는 모름

eBPF를 통해 packet, connection, latency 같은 **실시간 이벤트 기반 관측이 가능** → 이벤트 발생 순간 기록이 가능

```go
src: pod-A
dst: pod-B
event: tcp_retransmission
timestamp: 12:00:02
```

이를 통해 Pod A → Pod B 문제 임을 바로 알 수 있음

BUT 그렇다고 eBPF가 있다고 해서 상태 통계가 필요 없는 것은 아님

→ node saturation, kernel queue, conntrack 같은 노드 상태는 통계 메트릭이 더 좋다

**결론: 둘 다 필요하며 같이 보는 것이 좋음**

→ Prometheus에서 오는 통계 메트릭을 통해 이상 탐지 & Hubble에서 오는 이벤트를 통해 원인 분석

---

## Cilium이란

eBPF 기반 Kubernetes CNI(Container Network Interface)로, 워크로드 간 네트워크 연결(Networking)과 네트워크 관측 기능(Observability)을 모두 제공한다.

일반적인 CNI는 Linux 커널의 iptables를 기반으로 패킷을 처리한다.

반면 Cilium은 kube-proxy 및 iptables에 대한 의존도를 줄이고, Cilium Agent가 Kubernetes 상태를 기반으로 eBPF 프로그램을 생성하여 이를 커널에 로드함으로써 패킷 처리(라우팅, NAT, 로드밸런싱 등)를 수행한다.

```go
Pod -> iptables -> routing -> Pod
가 아닌
Pod -> eBPF program -> Pod
```

이를 통해 패킷 처리를 커널 내부에서 직접 수행할 수 있으며, iptables 기반 방식보다 더 낮은 지연과 높은 성능을 제공한다.

어떤 면에서 iptables가 느리고 커널로 하면 그게 왜 사라지는가?

### Cilium의 3계층 구조

Cilium은 크게 다음 3가지 계층으로 구성된다.

1. Control Plane: Cilium Agent
2. Datapath: eBPF Program
3. Observability: Hubble

---

### 1. Control Plane: Cilium Agent

Cilium Agent는 각 노드에서 실행되는 컨트롤 플레인 컴포넌트로, 클러스터의 NetworkPolicy와 Service를 관리하고 이를 기반으로 eBPF 프로그램을 구성한다.

**Kubernetes 상태 감시**

Cilium Agent는 Kubernetes API 서버를 watch하면서 다음과 같은 리소스를 감시한다.

- NetworkPolicy
- Service
- Endpoint/EndpointSlice
- Pod

이를 통해 클러스터 네트워크 상태 변화를 지속적으로 수신한다.

**eBPF 프로그램 관리**

수신한 정보를 기반으로

```yaml
Kubernetes API watch
↓
NetworkPolicy / Service 정보 수신
↓
eBPF bytecode 생성
↓
노드 커널에 로드 및 attach
```

과정을 수행한다.

또한, 노드 내 Pod 상태를 지속적으로 감시하며 필요한 네트워크 정책과 라우팅 정보를 커널 datapath에 반영한다.

→ 새로운 Pod가 생성되면 Cilium Agent가 해당 Pod의 인터페이스를 감지하고 그 즉시 해당 Pod에 적용될 정책을 계산하여 새로운 eBPF 바이트코드를 컴파일

즉, Cilium Agent는 클러스터 상태를 기반으로 eBPF datapath를 동적으로 구성하는 역할을 한다.

### 2. Datapath: eBPF Program

Datapath란 실제 패킷이 흐르는 경로에서 패킷을 처리하는 코드를 의미하고 Cilium에서 이는 Linux 커널에서 실행되는 eBPF 프로그램으로 구성된다.

Pod에서 발생한 네트워크 패킷은 커널 수준에서 eBPF 프로그램을 통과하며 다음 작업을 수행한다.

- NetworkPolicy 검사
- Service Load Balancing
- Routing
- NAT 처리
- 방화벽 필터링

이러한 처리는 커널 내부에서 수행되기 때문에 기존 iptables 기반 방식보다 컨텍스트 스위칭이 줄어들고 지연이 낮다.

또한 eBPF는 커널 재부팅 없이도 프로그램을 교체할 수 있어 Cilium Agent를 통해 네트워크 로직을 동적으로 업데이트할 수 있다.

**Flow Event 생성**

또한, eBPF datapath는 패킷 처리 과정에서 네트워크 flow 정보를 이벤트 형태로 생성할 수 있다. 

```yaml
source: podA
destination: kube-dns
protocol: UDP
port: 53
verdict: DROPPED
reason: Policy denied
```

이제 이러한 이벤트는 이후에 observability 계층에서 수집된다.

### 3. Observability: Hubble

Hubble은 Cilium의 네트워크 observability layer로, eBPF 프로그램에서 생성된 flow 이벤트를 수집하여 사용자에게 제공한다.

Hubble은 **패킷을 직접 캡처하는 도구가 아니다.**

Hubble은 다음과 같은 구조로 동작한다.

```yaml
eBPF datapath
↓
flow event 생성
↓
Cilium Agent
↓
Hubble collector
↓
CLI / UI / metrics exporter
```

즉,

- Cilium datapath → event 생성
- Hubble → event 수집 및 노출

의 관계를 가진다.

Hubble은 이 데이터를 CLI, Prometheus metrics exporter를 통해 활용할 수 있도록 제공한다.

---

### 전체 구조

정리하면 Cilium 아키텍처는 다음과 같이 동작한다.

```bash
Kubernetes API
↓
Cilium Agent (Control Plane)
↓
eBPF Program 설치
↓
Datapath에서 패킷 처리
↓
Flow Event 생성
↓
Hubble 수집
↓
Metrics / Flow Observability 제공
```