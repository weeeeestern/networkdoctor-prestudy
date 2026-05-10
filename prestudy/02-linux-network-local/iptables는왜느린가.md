# iptables는 왜 느린가..

Cilium은 기존 Kubernetes의 iptables 방식 대신 eBPF 프로그램을 활용하여 커널 내부에서 패킷 처리를 직접 수행함으로써, iptables 기반 방식보다 더 낮은 지연과 높은 성능을 제공한다.

**의문: 그럼 이때 iptables가 어떤 면에서 왜 느리고 커널 방식으로 하면 그게 왜 사라지는가?**

## iptables란 무엇인가

패킷을 검사하고 어떻게 처리할지 결정하는 규칙 엔진으로, 네트워크 패킷이 커널을 통과할 때 필터링, NAT, 라우팅 결정 같은 작업을 수행한다.

### 기본 역할

패킷이 네트워크 인터페이스(NIC)를 통해 들어오면 커널에서 다음과 같은 판단을 한다.

- 이 패킷을 허용할 것인가
- 어디로 보낼 것인가
- 주소를 바꿀 것인가(NAT)
- 차단할 것인가

iptables는 이러한 판단을 rule 기반으로 수행한다.

rule 기반으로 source IP 확인, destination port 확인 등을 거쳐 ACCEPT, DROP 등의 행동을 결정한다.

---

## Kubernetes와 iptables

### Kubernetes에서의 iptables rule 생성

Kubernetes에서 iptables rule이 생성되는 건 다음 4가지가 대표적이다.

- Service routing
- NodePort/ LoadBalancer
- Pod egress NAT
- NetworkPolicy

이때 Service routing과 NodePort/LoadBalancer 같은 Service 관련 rule은 kube-proxy가 담당하고, Pod networking과 policy rule은 CNI plugin이 담당한다.

1. **Service routing**

예를 들어, 다음과 같은 Service가 있으면

```yaml
Service IP: 10.96.0.10
Endpoints:
  Pod1
  Pod2
```

다음과 같은 iptables rule을 생성한다.

```yaml
10.96.0.10 → Pod1
10.96.0.10 → Pod2
```

1. NodePort/ LoadBalancer

예를 들어, 다음과 같은 NodePort 유형의 Service가 있다면

```yaml
NodeIP:30007 → Service → Pod
```

이거에 맞게 외부 트래픽을 해당하는 Service로 전달한다.

1. Pod egress NAT

Pod IP는 보통 cluster 내부 IP로 외부에서 라우팅 불가능하다

```yaml
Pod → Internet
↓
SNAT
↓
Node IP
```

따라서 Node IP로 바꾼다.

이게 SNAT rule이다. 이 rule이 CNI plugin에 의해 rule chain으로 생성된다.

1. NetworkPolicy 

예를 들어, 다음과 같은 NetworkPolicy가 있다면

```yaml
Pod A → Pod B 허용
Pod C → Pod B 차단
```

CNI plugin에 의해 해당하는 iptables rule로 구현된다.

이런 방식으로 실제 iptables는 Service 하나가 여러 chain을 만들고 Service 개수가 많아지면 rule이 폭발적으로 증가하게 된다.

### iptables의 패킷 처리 방식

iptables는 chain 기반 흐름으로 rule을 따라가며 패킷을 처리한다.

패킷이 NIC에서 들어오면 커널의 Netfilter hook에서 iptables가 실행된다.

대략적인 흐름은 다음과 같다.

```yaml
NIC
 ↓
PREROUTING
 ↓
routing decision
 ↓
INPUT / FORWARD
 ↓
POSTROUTING
```

각 단계마다 여러 chain이 존재하며 chain 내부에는 또 여러 rule이 있다.

```yaml
table
 ├─ chain
 │   ├─ rule
 │   ├─ rule
 │   └─ rule
```

rule에서는 다음과 같은 동작이 수행되며,

```yaml
ACCEPT # 패킷 처리 바로 끝남
DROP # 패킷 처리 바로 끝남
JUMP # 특정 chain으로 이동 -> jump한 chain에서 결정이 나지 않으면 원래 chain으로 돌아옴
DNAT
SNAT
```

rule의 JUMP를 통해 chain 간 이동이 이루어진다.

Kubernetes에서는 kube-proxy가 Service를 위해 여러 chain을 생성한다.

예를 들어 Service 요청이 들어오면 패킷은 다음과 같은 chain을 거친다.

```yaml
packet
 ↓
PREROUTING chain
 ↓
KUBE-SERVICES chain # 모든 서비스 entrypoint -> 서비스 매칭
 ↓
KUBE-SVC-XXXXX # 특정 Service -> 로드밸런싱
 ↓
KUBE-SEP-XXXXX # 특정 Pod -> DNAT (실제 목적지 IP로 변경)
 ↓
endpoint pod
```

이처럼 패킷은 rule 검사 → chain 이동→ rule 검사 → chain 이동의 흐름을 반복하며 처리된다.

이 흐름으로 패킷 처리가 이루어진다.

### iptables가 느려지는 이유

iptables의 rule 매칭은 기본적으로 순차 검사 방식이다.

즉, rule이 다음과 같이 있다면

```yaml
rule1
rule2
rule3
...
ruleN
```

패킷은 앞에서부터 rule을 검사한다.

따라서 rule 수가 증가할수록 검사 비용이 증가한다.

또한 Kubernetes에서는 Service와 Endpoint(Service backend pod)가 많아지면서 chain이 많이 생성되고 패킷이 여러 chain을 이동하며 rule을 검사하게 된다.

결과적으로 패킷 하나가 처리되는 동안 많은 rule + 많은 chain을 거치게 되며 지연이 증가할 수 있다.

### Cilium의 패킷 처리 방식

Cilium은 iptables 대신 eBPF 프로그램과 eBPF map을 사용한다.

iptables가 rule chain을 순차적으로 검사하는 방식이라면, Cilium은 map lookup을 통해 필요한 정보를 바로 조회한다.

예를 들어 Service 처리의 경우

```yaml
service_ip → backend pod list
```

와 같은 정보가 eBPF map에 저장되고,

패킷이 들어오면 eBPF 프로그램이 실행되고 필요한 정보를 map lookup으로 조회하여 바로 처리한다.

따라서 rule chain을 순차적으로 검사하는 iptables 방식과 달리 엔트리 수가 증가해도 lookup 비용이 크게 증가하지 않는다.