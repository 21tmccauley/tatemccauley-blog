---
layout: base.njk
title: "Medicine, Not Vitamins"
date: 2026-10-15
tags: post
excerpt: "I built a terminal UI to stop fighting our evidence-collection scripts every week. A sales engineer walked by, asked why customers didn't have it, and before long I was Paramify's first forward deployed engineer. What that taught me about building things people actually need."
---

# Medicine, Not Vitamins

Every Friday at Paramify we hosted free office hours for our customers to join and talk with us about how they could set up evidence automation with our platform. Evidence automation means pulling proof of your security controls straight from your systems, continuously, instead of gathering screenshots and spreadsheets once a year. It is a paradigm shift for so many companies, so we wanted to guide them on that journey. Each time, I saw their faces fall a bit when we explained that we had a series of scripts that could be run after a list of confusing commands in their terminal. This was not exactly the magic they were hoping to see.

The scripts our customers were wincing at were the same ones I was fighting with every week. As an intern, I was in charge of maintaining them and running them to collect evidence for our own environment. They were a collection of open-source scripts that funneled evidence from our tooling into our platform, but they were clunky: there wasn't great orchestration or a clear path to automated deployment. I like building dev tools, so I figured I'd make something to ease my own pain.

After some failed ideas and prototypes, I decided to make a TUI (a terminal user interface) to manage the scripts, easily store configs, create different manifests of which scripts to run when, and bundle it all up as a container. As a cherry on top, I made an ASCII "Paramify" logo and slapped it on top.

One time as I was using it, one of our sales engineers walked by and said, "Why are we not giving this version to our clients!?" I tried to explain that I had just made it for myself; it wasn't really made to share. Before I knew it, every demo was showing off our new tool for evidence automation.

Kenny Scott, Paramify's CEO, has a line he uses: people buy medicine, not vitamins. A vitamin is nice to have. You'll take it if it's in front of you, and you won't notice if you skip a day. Medicine is for something that hurts, and people will take it before it's finished, before it's pretty, before anyone has explained the dosage.

That sales engineer walking by was the moment the idea clicked for me. Nobody had asked for my TUI, and it wasn't on any roadmap. It spread because it touched something that hurt, for our customers and for us.

We knew customers needed out-of-the-box ways to configure, containerize and deploy their evidence collection pipelines. What I did not see coming was the demand for us to expand our library of scripts! I became the go-to person to explain how the tool worked and how to build new scripts and integrate them into this framework.

That demand changed the conversation. Our fetchers were already [open source](https://github.com/paramify/paramify-fetchers). Anyone could clone them, extend them, and point them at any platform. But having the code isn't the same as having it running in your environment, pulling the right evidence for the right requirements. Customers didn't want a repo. They wanted someone to come in and make it work.

So we started selling that: implementing and deploying the fetchers, building new ones, and growing the library along the way. Before I knew it, there was a new line item on the order form. It was me. I'd become a forward deployed engineer (FDE): an engineer who works inside a customer's environment to make the product actually work there.

Being first also meant there was no playbook. For a while, the FDE function was essentially me. There was no defined way to scope an engagement, no template for handing findings back to product, and a lot of the process lived in my head. We made it up as we went. As the team grew and the work became a real part of Paramify's services, the job shifted to writing it all down: formalizing how we scope, deliver, and hand off, so the next FDE wouldn't have to learn it the way I did.

My first engagement was a learning experience. Building tools I thought were interesting, to meet my own needs, was very different from being embedded in another company's process. I had to learn fast to look past what customers asked for and find the pain underneath.

One customer asked us to update all of our fetchers for their specific use case. What they really needed was a way to customize without cutting themselves off from the library, so we set them up with a private mirror: their changes on top, our updates still flowing in. Another customer asked for help writing control responses. The real problem was that they didn't have a clear picture of which tools and services were in their environment. So I ran every fetcher we had and looked at what came back. That data became the starting point for their control responses.

Before too long, every new customer was getting introduced to the Paramify TUI, and I was leading FDE engagements with several customers. Since July, the companies running it have ranged from AI labs and developer-tool startups to security vendors, HR and healthcare platforms, and defense and public-safety tech.

Not everything I built was medicine. As customer zero for our API, I pushed it harder than anyone, and I ended up building a full headless version of Paramify on top of it. I thought it was slick. But using our API was never anyone's problem. The pain was upstream: getting evidence out of their systems in the first place. The headless version taught me a lot about our API, but nobody was hurting without it.

If you're early in your career and trying to figure out where to put your effort, look for the thing everyone complains about and nobody owns: the scripts nobody wants to run, the process everyone works around. That pain is a map. Build something that makes it hurt less, even if it's just for yourself at first, and see who walks by and asks why they don't have it yet. That's how I went from an intern maintaining scripts to leading our FDEs on the technical side of engagements, while continuing to build out our program and the tools we sell to customers.
