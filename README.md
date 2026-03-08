# NetworkDoctor Pre-study

Pre-study repository for our capstone project: Kubernetes/Linux network troubleshooting, kernel metrics, and Go basics.

## Overview

This repository is a shared study space for our team before starting the main capstone project.

We use this repo to:
- investigate real Kubernetes network issue cases
- study Linux local network troubleshooting
- understand common kernel/network metrics used in practice
- learn Go language basics for future tooling and implementation
- upload weekly study results individually

## Study Topics

### 1. Kubernetes network issue cases
We investigate common network-related incidents in Kubernetes environments and summarize symptoms, causes, diagnostics, and possible solutions.

### 2. Linux network troubleshooting (local)
We practice local Linux network troubleshooting using commands, logs, interfaces, routing tables, DNS checks, and packet-level inspection.

### 3. Kernel metrics and practical monitoring cases
We study kernel-level and network-related metrics that are commonly collected in production environments, and learn how to interpret them.

### 4. Go language basics
We learn the fundamentals of Go syntax and programming patterns that can later be used for collectors, analyzers, or diagnostic tools.

## Repository Structure

```text
.
├─ docs/
├─ templates/
├─ 01-k8s-network-cases/
├─ 02-linux-network-local/
├─ 03-kernel-metrics/
├─ 04-go-basics/
└─ archive/
```
---
## 브랜치 규칙

- `main` 브랜치에는 직접 push하지 않습니다.
- 모든 작업은 각자 작업 브랜치를 만들어 진행합니다.
- 작업이 끝나면 Pull Request(PR)를 생성한 뒤 merge합니다.

### 브랜치 이름 규칙
```text
study/eunseo-week01

study/donguk-week01

study/seoyeong-week01
```
## 작업 방법

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