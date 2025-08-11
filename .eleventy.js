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

  // Add a shortcode for the current year
  eleventyConfig.addShortcode("year", () => `${new Date().getFullYear()}`);

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