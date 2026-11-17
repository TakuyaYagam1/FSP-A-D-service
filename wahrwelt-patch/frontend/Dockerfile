FROM nginx:1.31.8-alpine@sha256:fed7e14be83cbd2661a75cdc5b872d936cf155d46c327084d30b8df19ae88d72

RUN apk upgrade --no-cache

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --chmod=644 index.html login.html register.html styles.css app.js owned-cases.js route-pass.js auth.js route-map.svg logo.jpg favicon.svg /usr/share/nginx/html/

EXPOSE 80
