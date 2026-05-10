# NetworkDoctor Pre-study

NetworkDoctor 졸업 프로젝트를 시작하기 전, 팀원들이 공통 배경지식을 맞추기 위해 진행한 사전 스터디 자료입니다.

## Overview

이 폴더에는 프로젝트 본격 진행 전에 정리한 네트워크 관련 배경지식과 주차별 스터디 결과물이 모여 있습니다.

스터디에서 다룬 내용은 다음과 같습니다.

- Kubernetes 네트워크 문제 사례 조사
- Linux 로컬 네트워크 트러블슈팅
- 실무에서 사용하는 커널/네트워크 메트릭 이해
- 네트워크 토폴로지 구조 정리

## Study Topics
<img width="1320" height="762" alt="image" src="https://github.com/user-attachments/assets/ad368a50-f633-4351-aa68-ca8ea2bcc8d1" />


### 1. Kubernetes network issue cases
We investigate common network-related incidents in Kubernetes environments and summarize symptoms, causes, diagnostics, and possible solutions.

### 2. Linux network troubleshooting (local)
We practice local Linux network troubleshooting using commands, logs, interfaces, routing tables, DNS checks, and packet-level inspection.

### 3. Kernel metrics and practical monitoring cases
We study kernel-level and network-related metrics that are commonly collected in production environments, and learn how to interpret them.

### 4. Network topology
We study network topology structures and summarize how components are connected and communicate in practical environments.

## Directory Structure

```text
.
├─ 01-k8s-network-cases/
├─ 02-linux-network-local/
├─ 03-kernel-metrics/
└─ 05-network-topology/
```
---

## 당시 브랜치 규칙

- `main` 브랜치에는 직접 push하지 않습니다.
- 모든 작업은 각자 작업 브랜치를 만들어 진행합니다.
- 작업이 끝나면 Pull Request(PR)를 생성한 뒤 merge합니다.

### 브랜치 이름 규칙
```text
study/eunseo-week01

study/dongwook-week01

study/seoyoung-week01
```

## 당시 작업 방법

1. main 브랜치 최신 내용을 pull 받습니다. 
2. 자신의 작업 브랜치를 생성합니다. 
3. 해당 주차 폴더에 본인 파일을 작성합니다. 
4. commit 후 원격 브랜치로 push합니다. 
5. GitHub에서 PR을 생성합니다.

## PR 제목 Convention
[Week01] 은서 - Kubernetes 네트워크 문제 사례 조사

[Week01] 동욱 - 리눅스 로컬 네트워크 문제 분석

[Week01] 서영 - 커널 메트릭 수집 사례 정리

[Week01] 은서 - Go 기본 문법 정리
