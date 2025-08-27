---
layout: base.njk
title: "From Pipelines to Policies: Auditing Modern Application Development"
date: 2025-08-27
tags: post
excerpt: "My deep dive into auditing a modern CI/CD pipeline, connecting high-level security policies to the technical reality of developer workflows and tools like Snyk and SonarQube."
---

# From Pipelines to Policies: Auditing Modern Application Development

When I was assigned to the Maverik application change management audit, I’ll be honest, I had no idea what a "CI/CD pipeline" was. 

Over the course of two months, my task was to answer the question: "How do we make sure that new code, from a developer's keyboard to our live applications, is built and deployed securely?"

To find the answer, I couldn't just read a policy document. I had to learn what was actually happening in the development process, starting by interviewing the development teams themselves.

## The Digital Assembly Line

Through my conversations with developers, I learned that the CI/CD (Continuous Integration/Continuous Deployment) pipeline is like an automated assembly line for software. At Maverik, this assembly line was powered by a tool called **Jenkins**. A developer commits code, and Jenkins takes over, automatically building, testing, and preparing it for release.

My role was to walk the entire length of this assembly line. Each step of the process provided a level of assurance that the code was secure, had been reviewed by QA, and was ready to end up in production. I had to learn the entire application change management process from start to finish and see if each step was being followed correctly.

## From Conversation to Configuration Files

The most crucial part of the audit was verifying that security was truly built into the process, not just an afterthought. The developers told me they used a suite of tools like **Snyk** to scan for vulnerable dependencies and **SonarQube** to check for bugs and security issues in their own code.

But in an audit, "trust but verify" is the rule. It wasn't enough to hear that they were using these tools. I had to prove it.

This is where I had to roll up my sleeves and learn to read Jenkins configuration files. Sifting through code, I could directly verify the claims from my interviews. I looked for the specific lines that initiated the Snyk and SonarQube scans and checked the settings to confirm the pipeline would actually *fail* if a high-severity vulnerability was found. I was connecting the dots between a conversation with a developer and the technical enforcement in a config file.

## Connecting Code to the Committee

This technical deep dive was only one piece of the puzzle. I also learned how this automated process fed into the larger governance structure. Every significant change had to be approved by a **Change Advisory Board (CAB)**.

The logs and reports from the Jenkins pipeline—proof that the code was tested and passed all its security scans—served as the evidence presented to the CAB. I saw how a technical control, like a SonarQube scan configured in a Jenkins file, provided the assurance a high-level governance body like the CAB needed to make a risk-informed decision. I even got to see how this process was adapted for emergencies, ensuring security wasn't abandoned when things needed to move fast.

## My Key Takeaway

Auditing a CI/CD pipeline taught me that a modern auditor is a translator. You have to speak the language of developers and understand their tools, then translate that technical evidence into the language of risk and compliance for management and governance bodies. You are the bridge between the configuration file and the committee meeting. This experience proved that GRC isn't a separate world; it's a critical lens through which to view and validate the technology itself.

## What's Coming Next

**Next week**, I'll shift my focus from auditing the machines to testing the people. I'll share the story of how I was given permission to (ethically) phish my own company to test our human defenses.

---

*This is part of a series about my cybersecurity internship experiences. Catch up on [Post 1: The Two Sides of the Cybersecurity Coin](/posts/the-two-sides-of-the-security-coin/) and [Post 2: Auditing a Corporate Merger](/posts/what-i-learned-auditing-a-merger/).*