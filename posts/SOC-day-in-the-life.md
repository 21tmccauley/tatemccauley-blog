---
layout: base.njk
title: "The Digital Detective: A Day in the Life of a SOC Analyst"
date: 2025-09-19
tags: post
excerpt: "From the initial alert to the final verdict, a look at the investigative process of a SOC analyst hunting for threats in a sea of data."
---

# The Digital Detective: Investigating with Darktrace and Taegis XDR

Moving into the Security Operations Center (SOC) at Big West Oil meant shifting from theory to practice. My daily work revolved around two core platforms: **Taegis XDR** for alerts and endpoint data, and **Darktrace** for network visibility. My job was to be a digital detective, using these tools to investigate the constant stream of potential threats.

Every day brought a new set of puzzles. While each case was unique, the investigative process followed a clear, structured path.

## The First Clue: The Alert

An investigation always started with a single alert in Taegis XDR. I saw a wide variety of them, but the most common were related to user activity:
* **Potential Phishing Attempts:** A user account suddenly trying to access a known malicious domain.
* **Password Spraying:** A series of failed login attempts against multiple accounts from a single IP address.
* **Unusual Login Attempts:** A successful login from a geographic location where we have no employees.

These alerts were the first clue—the "tip-off"—that something was wrong. My job was to figure out if it was a real threat or just a false alarm.

## Building the Case File

Once an alert came in, I would immediately begin to build context around it. My investigation process involved three key questions:

1.  **Who and What is Involved?** Taegis XDR would show me the associated user and their device. The first step was understanding the context. Is this a server that should have very predictable behavior, or is it a developer's laptop with more varied activity?
2.  **Are There Other Clues?** I would then pivot to look for other alerts or events associated with that same user or device. A single unusual login might be a false positive. But an unusual login, followed by a "Potential Phishing" alert, followed by a "Malicious Process Detected" alert starts to paint a very clear picture of a compromise.
3.  **Is This Behavior Expected?** By combining these pieces of evidence, I could usually determine if the activity was expected or not. An engineer on vacation in Europe might trigger an unusual login alert, which is expected behavior. That same alert on a random Tuesday from an accountant's account is a major red flag.

## Delivering the Verdict

After gathering the evidence, I had to deliver a verdict. If the activity was deemed legitimate, I would document my findings and close the alert, sometimes tuning the rule to reduce future noise.

If it was a real threat, I would immediately escalate it to my manager. From there, we would work together to determine the next steps for containment and remediation. This hands-on process of finding the signal in the noise was the core of my role as an analyst. It was a thrilling challenge to piece together disparate logs and alerts to uncover the full story.

## What's Coming Next

Investigating alerts is one thing, but knowing which ones to investigate *first* is a completely different skill. **Next week**, I'll cover "The Art of Triage" and how a SOC analyst prioritizes threats when everything seems urgent.

---

*This is part of a series about my cybersecurity internship experiences. Read the [first post](/posts/the-two-sides-of-the-security-coin/) to understand the context of my journey through both GRC and Security Operations roles.*
```