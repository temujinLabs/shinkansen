server {
    listen 80;
    listen [::]:80;
    server_name shinkansen.temujinlabs.com;

    # Redirect all HTTP to HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name shinkansen.temujinlabs.com;

    # SSL certificates (managed by certbot / Let's Encrypt)
    ssl_certificate     /etc/letsencrypt/live/shinkansen.temujinlabs.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/shinkansen.temujinlabs.com/privkey.pem;
    include             /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;

    root /var/www/shinkansen;
    index index.html;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Cache static assets aggressively
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Cache HTML for a short time
    location ~* \.html$ {
        expires 1h;
        add_header Cache-Control "public, must-revalidate";
    }

    # Gzip
    gzip on;
    gzip_types text/plain text/css application/javascript text/html application/json image/svg+xml;
    gzip_min_length 256;

    # Serve static files, fall back to index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    access_log /var/log/nginx/shinkansen.temujinlabs.com.access.log;
    error_log  /var/log/nginx/shinkansen.temujinlabs.com.error.log;
}
