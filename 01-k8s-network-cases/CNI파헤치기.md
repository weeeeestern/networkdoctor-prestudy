## CNI란

컨테이너 런타임이 네트워크 설정을 **외부 플러그인에 위임**하기 위한 표준 인터페이스

즉, 

- 컨테이너 네트워크를 어떻게 구성할지를 CNI plugin이 결정
- 컨테이너 런타임(containerd)는 CNI를 호출만 함

+CNI는 컨테이너 네트워크 구성을 담당하며 실제 패킷 전달은 Linux 커널의 네트워크 스택(라우팅, iptables, eBPF 등)에 의해 처리된다.

### 컨테이너 생성 시 CNI plugin의 역할

1. 네트워크 네임스페이스 생성 → container runtime이 수행 후 CNI 호출
2.  veth pair 생성
    
    ```yaml
    # 한 쌍의 인터페이스 생성
    container eth0 <-> host veth
    ```
    
3. host쪽 `veth`를 네트워크에 연결
    - bridge 기반이면 → `cni0` bridge에 attach
    - calico면 → L3 routing(overlay or BGP) + iptables 사용
    - cilium이면 → eBPF datapath에 attach = 인터페이스에 eBPF hook을 건다
4. IP 할당
    - IPAM plugin이 각 인터페이스 쌍에 ip를 할당
5. container routing 설정
    - 같은 Pod 네트워크 대역은 `eth0`로 직접 통신
    - 외부로 나갈 땐 gateway (=bridge(`cni0` )의 IP)로 전달

→ 이 상태에서는

- 같은 노드 내 Pod 및 host와 통신 가능하며,
- 다른 노드 Pod 통신은 추가 routing 설정으로 Pod IP 그대로 통신 가능하고,
- 외부 통신을 위해서는 SNAT 설정이 필요하다.

### Pod(컨테이너)의 통신

1. **다른 노드 Pod 간**
- Pod 간 통신을 위해서는 각 노드가 다른 노드의 Pod CIDR을 알 수 있도록 라우팅 설정이 필요하며, 이는 CNI에 의해 자동으로 구성된다.
- 이는 정적인 규칙이 아니라 노드 및 Pod 상태에 따라 동적으로 관리된다.
1. **Pod → 외부**

외부로부터 응답이 정상적으로 들어올 수 있도록 NAT 테이블에서 외부로 나가는 패킷의 source IP를 Pod IP에서 Node IP로 변경하는 SNAT 규칙(postrouting)을 추가해야 한다.

이는 CNI 설치/초기화 시에 자동으로 구성되는데

- Calico, Flannel 등은 iptables POSTROUTING으로,
- Cilium은 eBPF에서 수행된다.

이 규칙은 **conntrack**이 기억하고 있다가 응답이 오면 host로 찍힌 dest 주소를 Pod IP로 다시 복원한다.

```yaml
container
 ↓
veth
 ↓
bridge (cni0)
 ↓
host routing 
 ↓
POSTROUTING SNAT (iptables or eBPF)
 ↓
internet
```

1. **외부 → Pod**

Pod IP는 private IP이므로, 외부에서 직접 접근할 수 없고, Service를 통해 노출해야 한다.

클러스터 내부 통신만 원한다면 ClusterIP, 외부 통신까지 원한다면 NodePort나 LoadBalancer 유형으로 생성하면 된다.

Service가 생성되면, 보통 **kube-proxy**가 이를 감지하고 iptables NAT table에서 DNAT 규칙을 추가하여 Service IP 또는 NodePort로 들어온 트래픽을 Pod로 전달한다.

```yaml
# NodePort
external → host:port
        ↓
PREROUTING DNAT
        ↓
container:port
```

이때, Cilium의 경우에는 좀 다르다.

Cilium의 경우에는 kube-proxy 없이 동작하며, iptables 대신 Cilium Agent가 Kubernetes 상태를 watch하여 eBPF map을 업데이트하거나 필요하면 eBPF 프로그램을 재로드하여 커널 datapath에 반영한다.