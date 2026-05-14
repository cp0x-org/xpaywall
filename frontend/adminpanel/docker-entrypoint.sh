#!/bin/sh
cat <<EOF > /usr/share/nginx/html/config.js
window.__CONFIG__ = {
  API_URL: '${API_URL:-http://localhost:9091/}',
  PROXY_URL: '${PROXY_URL:-http://localhost:8081/}',
};
EOF
exec nginx -g 'daemon off;'
