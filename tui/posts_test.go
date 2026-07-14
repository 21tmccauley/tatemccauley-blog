package main

import (
	"testing"
	"testing/fstest"
)

const samplePost = `---
layout: base.njk
title: "GRC Is Not Boring"
date: 2026-07-10
tags: post
excerpt: "An excerpt."
---

# GRC Is Not Boring

Body text.`

func TestSplitFrontMatter(t *testing.T) {
	meta, body := splitFrontMatter(samplePost)
	if meta == "" {
		t.Fatal("expected front matter, got none")
	}
	want := "# GRC Is Not Boring\n\nBody text."
	if body != want {
		t.Errorf("body = %q, want %q", body, want)
	}

	meta, body = splitFrontMatter("no front matter here")
	if meta != "" || body != "no front matter here" {
		t.Errorf("plain content: meta = %q, body = %q", meta, body)
	}
}

func TestParseMeta(t *testing.T) {
	meta, _ := splitFrontMatter(samplePost)
	title, date := parseMeta(meta)
	if title != "GRC Is Not Boring" {
		t.Errorf("title = %q", title)
	}
	if got := date.Format("2006-01-02"); got != "2026-07-10" {
		t.Errorf("date = %s", got)
	}
}

func TestLoadPostsSortsNewestFirstAndSkipsIndex(t *testing.T) {
	fsys := fstest.MapFS{
		"posts/old.md":   {Data: []byte("---\ntitle: \"Old\"\ndate: 2025-01-01\n---\nold body")},
		"posts/new.md":   {Data: []byte("---\ntitle: \"New\"\ndate: 2026-06-01\n---\nnew body")},
		"posts/index.md": {Data: []byte("---\ntitle: \"Listing\"\n---\nlisting page")},
		"posts/bare.md":  {Data: []byte("no front matter at all")},
	}
	posts, err := loadPosts(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 3 {
		t.Fatalf("got %d posts, want 3 (index.md skipped)", len(posts))
	}
	if posts[0].title != "New" || posts[1].title != "Old" {
		t.Errorf("order = [%s, %s], want newest first", posts[0].title, posts[1].title)
	}
	// No front matter: filename becomes the title, zero date sorts last.
	if posts[2].title != "bare" {
		t.Errorf("fallback title = %q, want %q", posts[2].title, "bare")
	}
	if posts[0].body != "new body" {
		t.Errorf("front matter not stripped from body: %q", posts[0].body)
	}
}
