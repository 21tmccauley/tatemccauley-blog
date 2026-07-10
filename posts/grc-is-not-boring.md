---
layout: base.njk
title: "GRC Is Not Boring"
date: 2026-07-10
tags: post
excerpt: "I spent two internships convinced GRC was the 'boring' side of security. Then I went to work on it full time at Paramify and found the opposite: real tools, real problems, and a field being rebuilt around AI. This is the start of a series on why."
---

# GRC Is Not Boring

Ask a room full of security people what they think of GRC, and you can watch the energy leave the room. Governance, risk, and compliance is the part of the field everyone pictures as spreadsheets, screenshots, and a checklist that lands in your inbox once a quarter. It's the work that sounds like homework. If you got into security to break systems or defend them, GRC is the corner you were quietly hoping to avoid.

A year ago, I would have nodded along.

In my first series on this blog, I wrote about seeing security from two sides: the auditor and the analyst, the building inspector and the firefighter. I cast GRC as the inspector. Important, but a step back from the action. Careful, methodical, and if I'm honest, a little dry.

Then I graduated and went to work on the inspector's side of the coin full time, building compliance tooling at Paramify. I expected paperwork. What I found was the most interesting engineering problem I've worked on.

## What the Work Actually Looks Like

My week is not spreadsheets. I build tools that reach into dozens of cloud accounts and pull back the evidence that proves a system is configured the way it's supposed to be. I've written agents that wake up on their own, notice a control has started failing, work out what changed, and write up why. The problems underneath are the same ones the rest of security cares about: handling credentials safely at scale, proving a system is actually in the state it claims to be in, and doing it continuously instead of once a quarter.

None of that is homework. It's distributed systems, automation, and increasingly AI. It just happens to live under a label most engineers wrote off years ago.

## Why Now

The timing isn't an accident either. Compliance is going through the biggest shift it's had in years, and most of the industry hasn't noticed yet.

For a long time, compliance meant working through a long list of prescriptive controls and writing a document for each one. FedRAMP 20x, the newest version of the U.S. government's cloud authorization program, throws a lot of that out. Instead of asking whether you wrote a policy, it asks whether you can actually demonstrate a security outcome in a way a machine can check.

That one change turns compliance from a writing exercise into an engineering one. If the answer has to be machine-checkable, someone has to build the thing that checks it. That someone increasingly looks like a software engineer, and more and more like an AI agent.

I'll spend a whole post on FedRAMP 20x soon, because it earns one.

## Why I'm Writing This

Partly because I think I'm early to something, and early is interesting. Partly because I'm learning as I go. GRC turned out to be a great place to learn, because the problems are real and the field is wide open.

Mostly, though, I want to change a few minds. If you've been avoiding this corner of security, I don't think you're seeing it clearly.

## What's Coming Next

Over the next few posts I'll make the case and show my work:

- why FedRAMP 20x is the most interesting thing happening in compliance, and what it actually changes
- the tools I get to build, and the real problems hiding inside the "boring" work
- where all of this goes as it becomes AI-native, and what I've learned trying to build for AI

---

*This kicks off a new series that picks up where [The Two Sides of the Cybersecurity Coin](/posts/the-two-sides-of-the-security-coin/) left off, this time from inside the GRC side of the coin.*