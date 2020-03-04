#!/bin/ash

if [ -z "$MF_MQTT_CLUSTER" ]
then
      envsubst '${MF_MQTT_ADAPTER_PORT}' < /etc/nginx/snippets/mqtt-upstream-single.conf > /etc/nginx/snippets/mqtt-upstream.conf
      envsubst '${MF_MQTT_ADAPTER_WS_PORT}' < /etc/nginx/snippets/mqtt-ws-upstream-single.conf > /etc/nginx/snippets/mqtt-ws-upstream.conf
else
      envsubst '${MF_MQTT_ADAPTER_PORT}' < /etc/nginx/snippets/mqtt-upstream-cluster.conf > /etc/nginx/snippets/mqtt-upstream.conf
      envsubst '${MF_MQTT_ADAPTER_WS_PORT}' < /etc/nginx/snippets/mqtt-ws-upstream-cluster.conf > /etc/nginx/snippets/mqtt-ws-upstream.conf
fi

envsubst '
    ${MF_USERS_HTTP_PORT}
    ${MF_THINGS_HTTP_PORT}
    ${MF_THINGS_HTTP_PORT}
    ${MF_HTTP_ADAPTER_PORT}
    ${MF_WS_ADAPTER_PORT}
    ${MF_UI_PORT}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf
    
if [ -n "$UI_INSIGHIO_PORT" ]; then
    sed -i -e "s/UI_INSIGHIO_PORT/$UI_INSIGHIO_PORT/" /etc/nginx/nginx.conf
else
    sed -i -e "s/UI_INSIGHIO_PORT/3004/" /etc/nginx/nginx.conf
fi
exec nginx -g "daemon off;"
