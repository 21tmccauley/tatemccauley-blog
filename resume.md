---
layout: base.njk
title: Resume
permalink: /resume/index.html
description: "Resume for Tate McCauley, a Forward Deployed Engineer at Paramify with a background in GRC, security operations, and IT administration."
---

# Tate McCauley

<div class="section headerInfo">

- tatemccauley@gmail.com
- 801-643-9827
- [linkedin.com/in/tate-mccauley/](https://linkedin.com/in/tate-mccauley/)
- [github.com/21tmccauley](https://github.com/21tmccauley)
- Provo, UT

</div>

## Education

### Brigham Young University, B.S.
Cybersecurity <span class="spacer"></span> Apr 2026

- **Scholarships:** Academic Scholarship (2024), Charles R. Bluemel Endowed Scholarship (2025)
- **Coursework:** Cloud Architecture, Secure Software Development, Penetration Testing, Web Development

## Experience

### Forward Deployed Engineer, Paramify <span class="spacer"></span> Feb 2026 &mdash; Present

- **Evidence Automation Framework:** Designed and built a schema-driven framework for automated compliance-evidence collection spanning **100+ integrations** (AWS, Azure, Okta, GitLab, Kubernetes, and more), with a single contract driving human, JSON/agent, and terminal-UI frontends and packaged for Docker, Kubernetes, and Terraform deployment.
- **Developer Tooling & SDK:** Authored an internal **Python SDK** and a git-style CLI (porcelain/plumbing commands, JSON-first output) that lets both engineers and AI agents drive the Paramify platform, shipped as versioned releases with self-contained binaries via CI.
- **AI-Native Automation:** Built agentic tooling that triages failing compliance validators unattended, running headlessly on a schedule to diagnose regressions and surface root causes, plus **MCP** servers providing semantic search and persistent memory over an internal corpus.
- **FedRAMP 20x:** Mapped the new **FedRAMP 20x** Key Security Indicators (KSIs) to concrete automated cloud checks across AWS, Azure, and GCP, translating the outcome-based baseline into machine-verifiable evidence.

### Image Team Technician (Security Engineering), BYU Office of Information Technology <span class="spacer"></span> Aug 2025 &mdash; Feb 2026

- **Automation Engineering:** Developed custom **PowerShell** automation to audit **Active Directory** inventory, programmatically flagging devices with incorrect naming conventions or misconfigured OUs to close security gaps.
- **Systems Integrity:** Manage the security posture of **1,000+ endpoints** using CrowdStrike Falcon, executing patch management cycles and maintaining system images to ensure enterprise compliance.
- **Infrastructure Maintenance:** Deploy software updates and maintain system integrity across a hybrid environment, ensuring high availability for library patrons and staff.

### Cybersecurity Analyst, Big West Oil <span class="spacer"></span> Apr 2025 &mdash; Nov 2025

- **Internal Tool Development:** Architected and deployed a custom SharePoint vulnerability tracking tool to replace static Excel workflows, leading remediation sprints every 10 days to drive a **58% reduction in compromised hosts**.
- **Event Monitoring:** Monitored high-volume security events (**450k+**) using SecureWorks Taegis XDR, filtering false positives and tuning detection logic to improve risk reporting accuracy.
- **Data Governance:** Conducted data access audits using **Varonis**, identifying and programmatically securing **3,500+ files** with improper public access permissions.

### IT/Cybersecurity Auditor, FJ Management <span class="spacer"></span> Jan 2025 &mdash; April 2025

- **Compliance & Infrastructure:** Audited complex firewall configurations (Palo Alto, Cisco) in a hybrid **IT/OT environment**; reviewed 200+ rules and removed **50 obsolete rules** to minimize the attack surface.
- **Strategic Integration:** Conducted post-acquisition security integration, analyzing redundant toolsets and presenting risk-based tech stack rationalization strategies directly to the CTO and CISO.
- **Social Engineering:** Engineered custom phishing campaigns targeting high-value roles to test human resilience and validate security training effectiveness.

### Computer and Server Technician, BYU Harold B. Lee Library <span class="spacer"></span> Jul 2023 &mdash; Dec 2024

- Managed OS and software updates for 30+ Linux (RHEL) and Windows Servers, ensuring system integrity and uptime.
- Performed weekly server backups and restoration testing while providing Tier 2 hardware support for library workstations.

## Key Projects

### Serverless Real-Time Chat Application
*Full-Stack Cloud Architecture (AWS, React, TypeScript)*
- **Full-Stack Engineering:** Built a responsive, type-safe client using **React 19**, **TypeScript**, and **Vite**, integrating **Radix UI** primitives for a modern, accessible user interface.
- **Event-Driven Architecture:** Engineered a real-time messaging platform using **AWS Lambda (Node.js)** and **API Gateway WebSockets**, handling persistent connections and message broadcasting.
- **Data Design & Optimization:** Designed a **DynamoDB** schema with Global Secondary Indexes (GSIs) to enable efficient, sorted message history retrieval while maintaining low-latency writes.
- **DevOps & IaC:** Provisioned 100% of the backend infrastructure using **Terraform** modules and wrote automated **Bash/PowerShell** scripts to bundle Lambda dependencies and deploy the frontend to **S3/CloudFront**.

### NetSTAR Shield (Capstone)
*Browser Security Extension (React, Python)*
- **Security Scoring Engine:** Developed a Python-based analysis engine that evaluates domains across 6 vectors (SSL/TLS, DNS, Email Security, HTTP Headers), utilizing a weighted harmonic mean algorithm to calculate real-time safety scores.
