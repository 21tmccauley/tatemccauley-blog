const { DateTime } = require("luxon");

module.exports = function(eleventyConfig) {
  // Add a filter for readable dates
  eleventyConfig.addFilter("readableDate", dateObj => {
    return DateTime.fromJSDate(dateObj, {zone: 'utc'}).toFormat("LLL dd, yyyy");
  });

  // Add a filter for ISO 8601 dates
  eleventyConfig.addFilter('isoDate', (dateObj) => {
    return DateTime.fromJSDate(dateObj, {zone: 'utc'}).toISODate();
  });

  // Add a filter for RFC 3339 datetimes (used by the Atom feed)
  eleventyConfig.addFilter('rfc3339', (dateObj) => {
    return DateTime.fromJSDate(dateObj, {zone: 'utc'}).toISO({ suppressMilliseconds: true });
  });

  // True for individual blog posts (not the /posts/ index) — drives article metadata
  eleventyConfig.addFilter('isPostUrl', (url) => {
    return typeof url === 'string' && url.startsWith('/posts/') && url !== '/posts/';
  });

  // Add a shortcode for the current year
  eleventyConfig.addShortcode("year", () => `${new Date().getFullYear()}`);

  // Copy static assets (OG image, etc.) straight through to the build
  eleventyConfig.addPassthroughCopy("assets");

  eleventyConfig.setFrontMatterParsingOptions({
    excerpt: true,
    excerpt_separator: ""
  });

  return {
    dir: {
      input: ".",
      includes: "_includes",
      output: "_site",
    }
  };
};