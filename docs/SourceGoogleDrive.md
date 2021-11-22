## Using a google drive as a JSONLD source

This involves using an API key, and an environment variable.
An enviroment variable is requeired, since multiple sources could use different api keys.
(and security... can't check in an env variable ;) )

(use case is Earthcube Resource Registry, where google form results are converted to jsonld files.)
```csv
"hack","SourceType","Active","Name","ProperName","URL","Headless","Domain","PID","Logo"
"51","googledrive","TRUE","ecrr_submitted","Earthcube Resource Registry","https://drive.google.com/drive/u/0/folders/1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd","FALSE","http://www.earthcube.org/resourceregistry/","","https://www.earthcube.org/sites/default/files/doc-repository/logo_earthcube_full_horizontal.png"
```

```yaml
- sourcetype: googledrive
  name: ecrr_submitted
  logo: https://www.earthcube.org/sites/default/files/doc-repository/logo_earthcube_full_horizontal.png
  url: https://drive.google.com/drive/u/0/folders/1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd
  headless: false
  pid: ""
  propername: Earthcube Resource Registry
  domain: http://www.earthcube.org/resourceregistry/
  active: true
  CredentialsEnv: GOOGLEAPIAUTH
  other: {}
```

The credentials files is located at:
```shell
setenv GOOGLEAPIAUTH = credentials/filename.json
```
## Credentials
A repository needs to provide a server_account json file
https://help.talend.com/r/E3i03eb7IpvsigwC58fxQg/ol2OwTHmFbDiMjQl3ES5QA

```json
{
  "type": "service_account",
  "project_id": "gleaner-",
  "private_key_id": "ke",
  "private_key": "-----BEGIN PRIVATE KEY----------END PRIVATE KEY-----\n",
  "client_email": ".iam.gserviceaccount.com",
  "client_id": "",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/gserviceaccount.com"
}
```