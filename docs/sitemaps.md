# Sitemaps

## About

Sitemaps can be in either TXT or XML format, with the XML being the more common.
Information about sitemaps can be found at https://www.sitemaps.org.

It is recommended that all data set landing pages be listed in a sitemap.

If your web site already has a sitemap, it is fine to add the URLs for 
your landing pages in that sitemap.   Systems like Gleaner and of course
Google and others will inspect pages for the desired JSON-LD data graph
packages.

## Sitemap Index

You can also use a sitemap index which is in effect a sitemap of sitemaps.  
An index might look like:

<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <sitemap>
      <loc>http://www.example.com/sitemap1.xml</loc>
      <lastmod>2004-10-01T18:23:17+00:00</lastmod>
   </sitemap>
   <sitemap>
      <loc>http://www.example.com/sitemap2.xml</loc>
      <lastmod>2005-01-01</lastmod>
   </sitemap>
</sitemapindex>

In this case sitemap1.xml might be your information and general site 
pages.  Then, sitemap2.xml could be dedicated to your data set landing pages.

A sitemap can only have 50,000 entries, so if you have more than that you will 
also need to use a sitemap index to spread the entries across 50K or less chunks 
with the files being referenced in the index.  

## Robots.txt

You can also list sitemaps in your robots.txt file and there are some 
interesting things you can do there as well to direct various agents to 
different results or sitemaps.  


Ref: https://tools.ietf.org/html/draft-rep-wg-topic-00

