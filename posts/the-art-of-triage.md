---
layout: base.njk
title: "The Art of Triage"
date: 2025-10-16
tags: post
excerpt: "When every alert looks urgent, how do you decide what's truly dangerous? A look at the most critical skill for a SOC analyst: triage."
---

# The Art of Triage: Prioritizing Threats in the SOC

In a busy Security Operations Center (SOC), the biggest challenge isn't just investigating alerts—it's knowing *which* ones to investigate first. When I started at Big West Oil, my Taegis XDR dashboard was a constant flood of information. If you try to treat every "High" severity alert with the same level of panic, you'll burn out in an hour and, worse, you'll miss the *truly* critical threat hiding in the noise.

This is the most essential and stressful skill of a SOC analyst: **the art of triage**.

It's not a simple checklist. It's a rapid-fire decision-making process that balances a few key questions to determine the answer to the most important one: "What do I have to work on *right now*?"



## 1. How Bad is the Threat? (Severity)

This is the most obvious starting point. My tools did a lot of the initial work for me. Taegis XDR would flag an alert as "High," "Medium," or "Low" based on the nature of the detection.

A "High" alert, like a potential ransomware signature or a known command-and-control (C2) callback, is almost always going to jump to the top of the list. A "Low" alert, like a single failed login, is less concerning on its own. But this is just the first piece of the puzzle.

## 2. Where is it Happening? (Asset Criticality)

This, I learned, is the most important question. An alert's true priority is a combination of its **Severity** and the **Criticality of the asset** it's on.

To be an effective analyst, you *must* understand the business. Is the alert on a critical domain controller or a server hosting the company's financial data? Or is it on a guest laptop on the public Wi-Fi?

A "Medium" severity alert on a domain controller is **infinitely more urgent** than a "High" severity alert on an intern's old laptop. This is where my GRC background from FJ Management gave me a massive head start. I already understood the "crown jewels" of the business. This let me see an asset as more than just an IP.

## 3. What's the Context? (Correlation)

No alert exists in a vacuum. A single, isolated alert is rarely a five-alarm fire. The real danger is in the *pattern*. This is where my investigative skills from the previous week came into play.

Is this "Unusual Login" alert an isolated event? Or was it preceded by a "Phishing Email Clicked" alert and followed by a "Malicious PowerShell Command" alert, all from the same user?

That chain of events tells a story. My job in triage was to quickly see if an alert was a standalone event or one chapter in a much scarier story. This is what triage is all about: using my tools and business knowledge to separate the signal from the noise, fast.

## The Final Decision

In just a few minutes, I had to make a call. Was this a false positive I could close? Was it a minor issue I could investigate when I had time? Or was this a true positive on a critical asset—a "drop everything and call my manager" moment?

Getting this right is the core of the job. It's what stops a small intrusion from becoming a front-page data breach.

## What's Coming Next

This daily act of balancing GRC context with SOC threats led to the biggest insight of my entire internship. **Next week**, I'll share that "Aha!" moment—how my audit experience made me a better analyst, and how my time in the SOC made me a better auditor.

---

*This is part of a series about my cybersecurity internship experiences. Read the previous posts to catch up on the journey so far.*
