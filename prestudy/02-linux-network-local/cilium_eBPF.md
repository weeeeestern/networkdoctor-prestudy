# eBPF

eBPF란, extended Berkeley Packet Filter의 준말로 커널 레벨에서 샌드박스가 적용된 프로그램을 실행할 수 있게 해주는 기술

- 기술이므로 linux 커널의 일부

커널 레벨 - 최고 권한 ( ↔ 유저 모드 )

샌드박스 - 안전장치로 적용된 프로그램 외의 시스템에 영향을 끼치지 않는 것

커널 내에서 가상 머신으로 패킷을 분석하고 필터링함 - 외부에 스택, 레지스터, 연산 등의 리소스를 제공하는 서버가 따로 있나?

:: 커널 안에 설계된 하드웨어적이 아닌 소프트웨어적인 가상의 CPU와 메모리 구조로 돌아감. JVM과 비슷한 결

![image.png](attachment:5fb05ddf-f122-4b01-93fc-d2103cdd860c:image.png)

자료구조(시스템)의 개선이 생기면서 기존 BPF에서 eBPF로 넘어감(해시 테이블 등 추가)

XDP라는 용어가 많이 보이는데, BPF의 프레임워크의 일종으로 고성능 패킷 처리를 가능하게 함
드라이버에서 패킷을 수신하는 순간 BPF 프로그램을 실행하므로 패킷 수신 즉시 라우팅 결정이 가능

동작 방식

![image.png](attachment:f92788d6-84b4-4cac-a8f7-6e98f59c8411:image.png)

전체적인 흐름을 보면

1. 어떤 프로그램이 프로세스를 진행
2. 이를 마치고 돌아가기 직전 eBPF 샌드박스 프로그램 실행
3. PID와 프로그램 이름을 가로채서 외부(유저 레벨)로 송신
4. 이후 os 스케쥴러
- 보안상의 이유(커널 레벨과 유저 레벨)로 직접 보내지 않고
- perf/ring buffer 통해 보냄

![image.png](attachment:20df6407-da81-4fec-a337-03369d5cc313:image.png)

kernel probe나 user probe를 이용해 eBPF 프로그램을 원하는 위치에 추가시킬 수 있음

그래서 eBPF를 편하게 적용하기 위해 cilium 사용!

어떤 툴을 어떻게 사용할지는 클로드가..

검증 스탭은 다음과 같다.

![image.png](attachment:dfd5fd1f-8ba4-474f-be6d-5bbbea33faa0:image.png)

1. eBPF 프로그램 실행이 가능한지 여부를 확인 후

2. 안전한지 확인

3. 이후 마지막으로 튜링 컴플릿 확인

![image.png](attachment:e2b41472-0f63-42e9-83dd-e75ba2167e2a:image.png)

구현 시작하면 보려고 가져옴

# cilium

eBPF 기반의 팟 네트워크를 구축하는 CNI Plugin

- 고성능 네트워킹 솔루션
- iptables을 이용한 쿠버네티스 트래픽 라우팅의 단점을 보완하여 네트워크 성능을 높히고자하는 목적
- Linux 컨테이너 관리 플랫폼을(Docker, Kubernetes) 사용하여 배포된 애플리케이션 서비스 간 네트워크 연결을 보호하는 오픈 소스 소프트웨어로 리눅스 자체에 강력한 보안 가시성(Security visibility)와 제어 로직을 동적으로 입력 가능
- Cilium은 속눈썹이라는 뜻으로 네트워크 레벨에서 연결을 보호하는 역할에 잘 들어맞음

ip table의 단점

- 규모가 커지면 latency가 기하급수적으로 커짐
- 각각 서로 다른 규칙을 통합해야 하기 때문
- 업데이트가 쉽지 않음
- 대규모 환경에서 ip table의 규칙을 이해하기가 어려움

→ cilium 등장@@

![image.png](attachment:2d25d11b-2470-4f8b-babc-2c484a121818:image.png)

위와 같이 크게 네 가지 구성요소로 이루어짐

- cilium
- hubble - 얘는 일종의 대시보드
- eBPF
- Data sources