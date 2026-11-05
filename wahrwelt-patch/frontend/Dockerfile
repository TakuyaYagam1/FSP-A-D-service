FROM nginx:1.31.8-alpine@sha256:a3554c0d0b48177b1413c26e01f695334428020d9df01243fbae3e11cfd17417

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --chmod=644 index.html login.html register.html styles.css app.js owned-cases.js route-pass.js auth.js route-map.svg logo.jpg favicon.svg /usr/share/nginx/html/

EXPOSE 80
