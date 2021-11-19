# Github 
You can use github as a repository for jsonLd files.
Put your jsonln files in a repository. Then you can use a github action to generate a sitemap and push to github pages.

Warning... if you use githubb pages, this will need to be modified. It can wipe out existing pages

To generate a sitemap using an action
1. create a directory: `.github/workflows`
2. add file `sitemap.yaml` from below: sitemap action workflow file
3. modify:
* with.base-url-path 
* with.path-to-root
4. commit
5. go to actions, see that actions run.
6. enable github pages
7. See if your sitemap is at location: https://{org}.github.io/{repo}/{path}/sitemap.xml

### sitemap action workflow file
This example publishes a sitemap for [earhtcube/GeoCODES-Metadata ](https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/) from the metadata/Datasets directory.

```
GeoCODES-Metadata
-- metadata
  -- Dataset
    -- jsonld files
```

We want to point at the raw Json files are on the main branch.go to  [metadata/dataset](ttps://github.com/earthcube/GeoCODES-Metadata/tree/main/metadata/Dataset), 
click on  a file, 
and select RAW, you will see  URL starting with

this is our **base-url-path**  `https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/`
our **path-to-root** is `metadata/Dataset`

For this Generated sitemap below, the sitemap url for a gleaner source will be:
https://earthcube.github.io/GeoCODES-Metadata/metadata/Dataset/sitemap.xml

```yaml
name: Generate xml sitemap

on:
  push:
    branches: [ main ]

jobs:
  sitemap_job:
    runs-on: ubuntu-latest
    name: Generate a sitemap

    steps:
    - name: Checkout the repo
      uses: actions/checkout@v2
      with:
        fetch-depth: 0

    - name: Generate the dataset sitemap
      id: sitemapdataset
      uses: cicirello/generate-sitemap@v1
      with:
        base-url-path: https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/metadata/Dataset/
        path-to-root: metadata/Dataset
        include-pdf: false
        additional-extensions: json jsonld
    - name: Output dataset stats
      run: |
        echo "sitemap-path = ${{ steps.sitemapdataset.outputs.sitemap-path }}"
        echo "url-count = ${{ steps.sitemapdataset.outputs.url-count }}"
        echo "excluded-count = ${{ steps.sitemapdataset.outputs.excluded-count }}"
    - name: push to gh pages
      uses: JamesIves/github-pages-deploy-action@4.1.6
      with:
        branch: gh-pages
        folder: .
```

add this to sources

```csv
"hack","SourceType","Active","Name","ProperName","URL","Headless","Domain","PID","Logo"
"52","sitemap","TRUE","ecrr_examples","Earthcube Resource Registry Examples","https://raw.githubusercontent.com/earthcube/ecrro/master/Examples/sitemap.xml","FALSE","http://www.earthcube.org/resourceregistry/examples","","https://www.earthcube.org/sites/default/files/doc-repository/logo_earthcube_full_horizontal.png"
"","sitemap","TRUE","geocodes_examples","GeoCodes Tools Examples","https://raw.githubusercontent.com/earthcube/GeoCODES-Metadata/main/sitemap.xml","FALSE","https://github.com/earthcube/GeoCODES-Metadata/","",""
```