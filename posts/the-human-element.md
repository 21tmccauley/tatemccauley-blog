---
layout: base.njk
title: "The Human Element: I Phished My Own Company (For Science!)"
date: 2025-09-03
tags: post
excerpt: "How I used OSINT, AI, and a little social engineering to build a realistic spearphishing campaign to test the human firewall at my own company."
---

# The Human Element: I Phished My Own Company (For Science!)

After auditing the machines and software pipelines, my focus shifted to the most complex and unpredictable part of any security system: people. Upper management at FJ Management wanted to know how our teams would hold up against a sophisticated, targeted spearphishing attack. My manager gave me the project, and I was given permission to (ethically) attack my own companies.

The goal wasn't to play "gotcha" or to get anyone in trouble. It was to gather data and see how we could build a stronger, more resilient security awareness culture.

## Step 1: Building the Attack Plan with OSINT

To make this a true test, we decided to use as little inside information as possible. The goal was to simulate what a real attacker would do. That meant my first step was **Open-Source Intelligence (OSINT)** gathering.

Using LinkedIn profiles, I pieced together an organizational chart for our target groups at FJ Management and its subsidiaries. I identified key individuals, their job titles, and their roles within the company. This public information would become the foundation for creating highly specific and believable phishing lures.

## Step 2: Crafting the Bait with AI

With our targets identified, it was time to craft the bait. We used the open-source tool **Gophish** to manage the campaigns. For the emails and landing pages, I turned to AI.

Based on the job titles I found, I created **15 bespoke and targeted emails**. For example, an email to someone in finance might reference a fake update for an invoice, while a message to an IT-adjacent role might be about another user who is having issues. The AI helped me write convincing copy for each scenario.

To make the attacks even more realistic, we registered several **typosquatted domains** that looked nearly identical to our real ones—a classic trick where a capital 'I' can look just like a lowercase 'l', making the fake domain almost impossible to spot at a glance. These domains hosted AI-designed fake landing pages that were nearly identical clones of a SharePoint or Microsoft 365 login portal.

## Step 3: The Launch

I’ll admit, it was nerve-wracking to hit "send." These weren't generic, typo-filled phishing emails. They were carefully crafted attacks aimed at my own colleagues. But I reminded myself that we were doing this to help protect the company in the long run we were a sparring partner, not a real opponent.

We started seeing results almost immediately. People clicked.

Even many who didn't click reached out to our security team to comment that it was the most convincing phishing attempt they'd ever encountered. That feedback alone was a huge win—it meant our simulation was realistic.

## The Surprising Results

Here’s where our test differed from a standard phishing campaign. When an employee clicked the link and landed on our fake page, there was no obvious "You've Been Phished!" message. We tracked the click, but we didn't harvest credentials. Our primary goal was to see what happened *next*. **Would they report it?**

We found that while employee training was generally effective, the combination of specific details gathered from OSINT and the polished, AI-generated landing pages gave us a significant degree of success. Some people who clicked never realized that it was a spearphishing attack. Others did realize after they entered their credentials but were hesitant to report their mistake to the security team.

My key takeaway was that the "human firewall" is not a pass-fail system. It's a dynamic defense that needs to be tested with realistic, modern threats. By acting like a real attacker, we gained invaluable data on how to better train and equip our team for the real thing.

## What's Coming Next

My time in the GRC world, from auditing mergers to testing our human defenses, gave me a strategic view of security. **Next week**, I'll describe my transition to the other side of the coin: moving to the front lines as a Cybersecurity Analyst at Big West Oil.

---

*This is part of a series about my cybersecurity internship experiences. Read the previous posts to catch up on the journey so far.*






