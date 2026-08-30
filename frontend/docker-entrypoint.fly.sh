#!/bin/sh
set -e

UPSTREAM="${API_UPSTREAM:-onbober-api.internal:8080}"
sed "s|__API_UPSTREAM__|${UPSTREAM}|g" \
  /etc/nginx/templates/default.conf.template \
  > /etc/nginx/conf.d/default.conf

exec nginx -g 'daemon off;'
