#!/bin/bash
# gnu-date - a wrapper script for invoking `date(1)` from within a Docker image
docker run --rm -ti earthcube/gleaner:latest  /gleaner  "$@"
