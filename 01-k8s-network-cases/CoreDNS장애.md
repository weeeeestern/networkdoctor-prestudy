# CoreDNS 장애

주요 발생 계층: Kubernetes / Network

클러스터 DNS 구성 또는 네트워크 경로에서 지연 발생

---

## DNS 이상 징후 감지

CoreDNS는 Prometheus metrics 플러그인을 통해 DNS 요청 수, 응답 코드, 처리 시간 등의 메트릭을 노출한다.

```bash
.:53 {
    errors
    health
    kubernetes cluster.local
    prometheus :9153 # metrics 플러그인
    forward . /etc/resolv.conf
    cache 30
}
```

이와 같이 내부 상태에 대한 데이터를 `/metrics` 엔드포인트로 노출하며, Prometheus는 이 엔드포인트를 scrape하여 DNS 관련 메트릭을 수집한다.

### 대표 메트릭

```bash
coredns_dns_requests_total{server, zone, view, proto, family, type} # total query count
coredns_dns_request_duration_seconds{server, zone, view, type} # duration to process each query
coredns_dns_responses_total{server, zone, view, rcode, plugin} # response per zone, rcode and plugin.
```

이러한 메트릭을 기반으로 

- DNS 요청 폭증 → 어플리케이션 또는 DNS 설정 문제
- p99 DNS latency 증가 → 일부 DNS 요청 처리 시간이 크게 증가
- servfail 증가 → CoreDNS 내부 처리 실패

같은 이상 징후를 탐지할 수 있다

## DNS 쿼리 흐름 확인

이상 징후가 감지되면 다음 단계는

```bash
Application Pod
↓
Pod /etc/resolv.conf
↓
CoreDNS Service
↓
CoreDNS Pod
↓
Kubernetes API
↓
DNS 응답
```

이와 같은 DNS 요청 흐름에서 DNS 요청이 어느 단계까지 도달했는지 확인하는 것이다.

DNS 로그는 CoreDNS ConfigMap에 `log` 플러그인이 활성화되어 있을 때 확인할 수 있다.

1. 로그가 정상적으로 보임 → 문제 x
    
    ```bash
    172.17.0.18:41675 - "A IN kubernetes.default.svc.cluster.local." NOERROR
    ```
    
    DNS 자체 문제일 가능성은 낮음
    
2. 로그가 전혀 없다 → CoreDNS에 도달하지 못함
    
    CoreDNS 이전 단계 문제임
    
    원인 후보
    
    - Pod DNS 설정 문제 → `/etc/resolv.conf` 설정 문제
        
        ```bash
        kubectl exec pod -- cat /etc/resolv.conf
        ```
        
        - ndots 설정 등으로 인해 같은 요청도 먼저 내부 dns로 여러 번 질의함으로써 dns 요청이 늘어날 수 있음
            
            ```bash
            search default.svc.cluster.local svc.cluster.local cluster.local google.internal c.gce_project_id.internal
            nameserver 10.0.0.10
            options ndots:5
            ```
            
        - ex: api.example.com 요청 시
            
            ```bash
            api.example.com.default.svc.cluster.local
            api.example.com.svc.cluster.local
            api.example.com.cluster.local
            api.example.com
            ```
            
            ⇒ 한 번의 DNS lookup → 여러 DNS query 발생
            
    - CoreDNS Pod / Service 문제
    - CNI 네트워크 문제
    - NetworkPolicy 문제
3. 로그는 있는데 SERVFAIL
    
    ```bash
    "A IN serverproxy.contoso.net.cluster.local." SERVFAIL
    ```
    
    CoreDNS가 요청은 받았지만 처리 실패
    
    원인 후보
    
    - CoreDNS의 Kubernetes API 권한 문제
        
        CoreDNS는 Service를 찾기 위해 service, endpoint, endpointslice를 조회하는데, 이때 적절한 권한이 없으면 실패 → ClusterRole 확인 필요
        
    - CoreDNS 리소스 문제
        
        CPU saturation, memory pressure
        
    - conntrack saturation
        - DNS 요청이 많은 경우 conntrack table full이 발생할 수 있음

---

참고 문헌

[Kubernetes | Debugging DNS Resolution](https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/#known-issues)

https://coredns.io/plugins/metrics/