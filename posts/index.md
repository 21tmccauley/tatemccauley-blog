---
layout: base.njk
title: Blog
permalink: /posts/index.html
description: "Essays by Tate McCauley on security, compliance, and building AI-native tools — from SOC work and IT audits to FedRAMP and GRC engineering."
---

## Blog

<p class="muted">{{ collections.post | size }} posts, newest first</p>

<ul class="post-list">
{%- assign newest = collections.post | reverse -%}
{%- for post in newest -%}
  <li class="post-entry">
    <p class="post-meta">
        <time datetime="{{ post.date | isoDate }}">{{ post.date | readableDate }}</time>
    </p>
    <a href="{{ post.url | url }}" class="post-title">{{ post.data.title }}</a>
    {%- if post.data.excerpt %}
    <p class="post-excerpt">{{ post.data.excerpt }}</p>
    {%- endif %}
  </li>
{%- endfor -%}
</ul>
