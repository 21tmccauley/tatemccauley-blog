---
layout: base.njk
title: Home
permalink: /
---

## Hey, I'm Tate.

I'm a cybersecurity student at BYU with a passion for understanding how our digital world is built, connected, and defended. This is  where I write about my journey in tech, security, and other things I find interesting.

You can learn more [about me](/about/), view my [resume](/resume/), or see what I'm focused on [right now](/now/).

---

## Recent Posts

<ul class="clean-list no-underline bold-links">
{%- for post in collections.post | reverse | limit(3) -%}
  <li>
    <a href="{{ post.url | url }}">{{ post.data.title }}</a>
    <time class="muted" datetime="{{ post.date | isoDate }}">{{ post.date | readableDate }}</time>
  </li>
{%- endfor -%}
</ul>

[View all posts...](/posts/)