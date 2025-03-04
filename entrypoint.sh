#!/bin/sh
set -x
yq -i ".domainName = \"${DOMAIN_NAME}\"" /app/noProxy.yaml
exec $@