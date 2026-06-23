FROM nginx:alpine
COPY dist/demo-ui.iife.js /usr/share/nginx/html/apps/demo-ui/demo-ui.js
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
