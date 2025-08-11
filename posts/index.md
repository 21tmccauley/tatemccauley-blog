---
layout: base.njk
title: Blog
permalink: /posts/index.html
eleventyExcludeFromCollections: true
---

## Blog

<ul class="post-list">
{%- for post in collections.post | reverse -%}
  <li class="post-entry">
    <a href="{{ post.url | url }}" class="post-title">{{ post.data.title }}</a>
    <p class="post-meta">
        <time datetime="{{ post.date | isoDate }}">{{ post.date | readableDate }}</time>
    </p>

    {% if post.data.excerpt %}
        <p class="post-excerpt">
            {{ post.data.excerpt }}
        </p>
        <a href="{{ post.url | url }}" class="read-more">Read more...</a>
    {% endif %}

  </li>
{%- endfor -%}
</ul>