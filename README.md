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