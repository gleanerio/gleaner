## Using a Sitemap as Source
This is fairly simple.
* name, propername
* Url of the sitemap,
* select an identifier as a PID, usually link to a repository of repositories, or a doi
* domain (WHAT IS THE DOMAIN)
* headless: if the repository generates the JSON dymically in the client, then true.
* active: true if you want this to be utilized in when running. We have  a list of all partners, but all partners to not have JSONLD or sitemaps. 

```csv
"hack","SourceType","Active","Name","ProperName","URL","Headless","Domain","PID","Logo"
"3","sitemap","TRUE","opentopography","OpenTopography","https://opentopography.org/sitemap.xml","FALSE","http://www.opentopography.org/","https://www.re3data.org/repository/r3d100010655","https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png"
```
A standard sitemap is below:
```yaml
sources:
  - sourcetype: sitemap
      name: opentopography
      logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
      url: https://opentopography.org/sitemap.xml
      headless: false
      pid: https://www.re3data.org/repository/r3d100010655
      propername: OpenTopography
      domain: http://www.opentopography.org/
      active: true
```
