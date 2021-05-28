#!/bin/bash
# init.d service script

INSTALL_DIR="/usr/local/scheck"
SERVICE=scheck
BINARY="${INSTALL_DIR}"/scheck
CFG="${INSTALL_DIR}"/scheck.conf
PID="${INSTALL_DIR}"/scheck.pid
LOG="${INSTALL_DIR}"/log

start() {

    if [ -f "$PID" ] && kill -0 $(cat "$PID"); then
        echo 'Service already running'
        return 1
    fi
    echo 'Starting service…'
    ${BINARY} -config "${CFG}" & # run in backend
    if [ $? -eq 0 ] ;then
        echo "$!" > "${PID}"
        echo 'Service started'
    else
        echo "failed"
    fi
}

stop() {
    if [ ! -f "$PID" ] || ! kill -0 $(cat "$PID"); then
        echo 'Service not running'
        return 1
    fi
    echo 'Stopping service…'
    kill -15 $(cat "$PID") && echo -n "" > "$PID"
    echo 'Service stopped'
}

status() {
    if [ -f "$PID" ] && kill -0 $(cat "$PID"); then
        echo 'Service is running'
    else
        echo "$SERVICE stopped"
    fi
}

restart() {
    stop
    echo "Restarting ${SERVICE}..."
    sleep 1
    start
}

case "$1" in
    start|stop|restart|status)
        $1
        ;;
    *)
        echo $"Usage: $0 {start|stop|restart|status}"
        exit 1
    ;;
esac
