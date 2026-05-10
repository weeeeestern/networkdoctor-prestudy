# Network Policy 로 인한 Drop 탐지

## 1. NetworkPolicy란

Kubernetes에서 Pod 간 또는 Pod와 외부 간의 네트워크 트래픽을 제어하기 위해 NetworkPolicy를 사용한다.

NetworkPolicy는 IP/Port/Pod/Namespace 기준으로 트래픽을 허용하거나 차단하는 정책이며 어플리케이션 단위의 네트워크 제어를 가능하게 한다.

예시 정책

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: default
spec:
  podSelector:
    matchLabels:
      role: db
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - ipBlock:
        cidr: 172.17.0.0/16
        except:
        - 172.17.1.0/24
    - namespaceSelector:
        matchLabels:
          project: myproject
    - podSelector:
        matchLabels:
          role: frontend
    ports:
    - protocol: TCP
      port: 6379
  egress:
  - to:
    - ipBlock:
        cidr: 10.0.0.0/24
    ports:
    - protocol: TCP
      port: 5978
```

## 2. NetworkPolicy의 기본 동작

Pod는 Ingress와 Egress 두 가지 방향의 트래픽 격리를 가진다.

기본 상태 

- NetworkPolicy 없음 → 모든 트래픽 허용

하지만 특정 Pod에 대해 NetworkPolicy가 하나라도 적용되면

- default allow → default deny

로 바뀐다.

즉, 허용 규칙에 명시된 트래픽만 통과하고 나머지는 차단된다.

## 3. NetworkPolicy 구현 방식

NetworkPolicy는 Kubernetes가 아닌 CNI 플러그인에 의해 구현된다.

Cilium은 iptables가 아닌 eBPF 기반의 대표적인 CNI이다.

Cilium은 현재 L3/L4를 지원하는 표준 NetworkPolicy와 더불어 L3-L7까지 지원하는 확장된 형태의 CiliumNetworkPolicy까지 CRD로 제공한다.

## 4. NetworkPolicy 문제 탐지

### Cilium

NetworkPolicy 문제는 패킷 drop 이벤트를 통해 탐지할 수 있다.

Cilium은 패킷이 차단될 때 drop reason을 메트릭으로 노출한다.

```yaml
# 총 손실 패킷 수
drop_count_total{reason,direction}
```

ex

```yaml
cilium_drop_count_total{
  reason="Policy denied"
}
```

값이 증가하면 → NetworkPolicy에 의해 패킷이 drop되고 있다는 것을 알 수 있음

### Hubble Flow 기반

Cilium은 패킷 흐름 정보를 관측하는 observability 도구인 Hubble을 제공한다.

Hubble은 다음과 같은 네트워크 flow 이벤트를 생성한다.

```yaml
# podA -> kube-dns 가 NetworkPolicy에 의해 차단됨
source: podA
destination: kube-dns
verdict: DROPPED
policy: deny
```

이러한 flow 이벤트는 Hubble metrics exporter를 통해

```yaml
source: podA
destination: kube-dns
verdict: DROPPED
policy: deny
```

이와 같이 Prometheus 메트릭으로 변환하여 수집할 수 있다.

---

만일 NetworkPolicy에 의해 DNS가 차단된 경우, 다음과 같은 방식으로 추론해나갈 수 있다.

1. CoreDNS request 감소 → 이상 탐지
    
    ```yaml
    coredns_dns_requests_total
    ```
    
    Pod에서 DNS 요청이 줄어들면 CoreDNS에 들어오는 요청도 감소
    
2. Cilium drop 이벤트 증가
    
    ```yaml
    cilium_drop_count_total{
      reason="Policy denied"
    }
    ```
    
    NetworkPolicy로 인해 패킷이 차단되고 있음을 의미
    
3. Hubble flow에서 DNS drop 확인
    
    ```yaml
    destination = kube-dns
    verdict = DROPPED
    ```
    
    Pod → CoreDNS 로 가는 DNS 요청이 policy에 의해 drop
    

종합하여 Pod의 egress NetworkPolicy에 CoreDNS 허용 규칙이 없는 것이 원인임을 추론할 수 있다.

---

참고 문헌

https://kubernetes.io/docs/concepts/services-networking/network-policies/

https://docs.cilium.io/en/stable/network/kubernetes/policy/#ciliumnetworkpolicy

https://docs.cilium.io/en/stable/observability/metrics/#cilium-feature-network-policies