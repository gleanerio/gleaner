# Review of gleaner bucket structure

Do I validate as well formed (not SHACL valid) the JSON-LD as valid JSON-LD before writing to summoned bucket?

Milled object are only used by me to load via nabu

Prov used by me (via s3select) to know what I have 

results object might be redundant if I build orgname_runid_date.parquet pattern for master milled documents

Keep JSON-LD in summoned

NOTE:  parquet for milled and prov  (and for orgs?)   These are really used only by me and as 
such do impact others.  

orgs:  keep each parquet...   then remove and rebuild a caoncat them to a master org file each run



nas/gleaner
├─ milled
│  ├─ org1
│  └─ org2
├─ orgs   (individual files)
├─ prov
│  ├─ org1
│  └─ org2
├─ results
│  ├─ runid1
│  └─ runid2
├─ shapes  (individual input files...   why in object store?  why does Gleaner do ths?  Better done elsewhere...)
├─ summoned
│  ├─ org1
│  └─ org2
└─ verified
   ├─ org1
   └─ org2
