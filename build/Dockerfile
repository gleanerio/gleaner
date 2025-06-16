# Start from scratch image and add in a precompiled binary
# CGO_ENABLED=0 env  go build .
# docker build  --tag="opencoredata/ocdweb:0.9.4"  .
# docker run -d -p 9900:9900  opencoredata/ocdweb:0.9.4
FROM alpine
#FROM scratch

#RUN apk add py-urllib3 openssl certbot curl --no-cache \
RUN apk add openssl certbot curl --no-cache \
    --repository https://dl-3.alpinelinux.org/alpine/v3.7/community/ \
    --repository https://dl-3.alpinelinux.org/alpine/v3.7/main/ \
    && rm -rf /var/cache/apk/*

# Add in the static elements (could also mount these from local filesystem)
RUN mkdir /nabu
RUN mkdir /nabu/config
RUN mkdir /nabu/assets
ADD ./cmd/nabu/nabu /nabu/
ADD ./assets /nabu/assets

RUN mkdir /assets
# just in case thin
#ADD ./assets/schemaorg-current-https.jsonld /assets/schemaorg-current-https.jsonld
#ADD ./assets/schemaorg-current-http.jsonld /assets/schemaorg-current-http.jsonld
RUN curl https://schema.org/version/latest/schemaorg-current-https.jsonld --output /assets/schemaorg-current-https.jsonld
RUN curl https://schema.org/version/latest/schemaorg-current-http.jsonld --output /assets/schemaorg-current-http.jsonld

# ADD ../web/static  /static   # Replace with -v mounting the /web/static directory
# static is in the .dockerignore..  so an emptry dir is made unless static is removed
# from the ignore file... we mount the volume via compose from the local FS

# Add our binary
ENTRYPOINT ["/nabu/nabu"]
