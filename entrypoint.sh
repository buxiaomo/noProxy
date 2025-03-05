#!/bin/sh
set -x
if [ -n "$DOMAIN_NAME" ]; then
  yq -i ".domainName = \"${DOMAIN_NAME}\"" /app/noProxy.yaml
fi
exec $@