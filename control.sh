#!/bin/bash
# ======================灵力平台上线需要===================================#
#-----------------------文件及目录定位-------------------------------------#
basename=$(basename "$0")
# get curr path
if [ -L "$0" ]
then
    file=$(readlink -f "$0")
else
    file=$0
fi
ACTS_DIR=$(cd $(dirname $file); pwd)
#------------------------------------------------------------#

#---------------------自定义配置区---------------------------#
#程序名称
APP_NAME="socketserver"
#设置系统环境变量
export GDP_MODE=release
#执行命令
CMD="${ACTS_DIR}/${APP_NAME}"
# 增加加载库的地址
# export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${ACTS_DIR}/libs/pta/lib

#------------------------------------------------------------#

PORT_MAIN=$(echo $MATRIX_INSTANCE_PORTS | tr , '\n' | awk -F'=' '$1=="main"{print $2}')
IDC_MAIN=$(echo $MATRIX_INSTANCE_TAGS | tr , '\n' | awk -F'=' '$1=="isp"{print $2}')

ACTS="$APP_NAME"
SUPERVISE="$ACTS_DIR/supervise/bin/supervise"
SUPERVISE_CONF="$ACTS_DIR/supervise/conf/supervise.conf"
STATUS_DIR="$ACTS_DIR/supervise/status/$ACTS"


SIGNAL_SUP=-SIGTERM
SIGNAL=-SIGTERM

#parse option
set -- $(getopt f:C: "$@")
while [ $# -gt 0 ]
do
    case "$1" in
        (--) shift; break;;
        (-*) echo "Unrecognized option $1";;
        (*)  echo "$1"; break;;
    esac
    shift
done

STATUS_FILE=$STATUS_DIR/status
LOCK_FILE=$STATUS_DIR/lock
CONTROL_FILE=$STATUS_DIR/control

_message() {
    echo "$@"
}

_warning() {
    _message WARNING: "$@"
}

is_running() {
    #/sbin/fuser $LOCK_FILE &>/dev/null
    ps axo pid,cmd | grep "${SUPERVISE}\>" | grep -v "grep" &>/dev/null
}

get_ppid() {
    #/sbin/fuser $LOCK_FILE 2>&1 | awk '{print $2}'
    ps axo pid,cmd | grep "${SUPERVISE}" | grep -v "grep" | awk '{print $1}'
}

get_pid() {
    od -d --skip-bytes=16 $STATUS_FILE | awk '{print $2}'
}

_kill() {
    kill -9 "$@" 2>/dev/null || true
}

_killgroup() {
    killall -9 supervise ${APP_NAME} || true
    #kill -TERM -"$1" 2>/dev/null || true
}

control_start() {
    AR_APP_TOMAL="${ACTS_DIR}/conf/ws_server.toml"
    sed -i "s/\(listen_addr[^\"]*=[^\"]*\).*$/\1\"0.0.0.0:${PORT_MAIN}\"/g" AR_APP_TOMAL
    sed -i "s/\(idc[^\"]*=[^\"]*\).*$/\1\"${IDC_MAIN}\"/g" AR_APP_TOMAL
    if is_running
    then
        _warning "$ACTS is already running"
        exit 1
    fi

#    sleep 2

    if [ ! -n "$CMD" ];then
        _warning "param CMD is not defined!"
        exit -1
    fi

    [ -d "$STATUS_DIR" ] || mkdir -p "$STATUS_DIR"
    $SUPERVISE -f "$CMD" -p "$STATUS_DIR" -F "$SUPERVISE_CONF" &>/dev/null </dev/null &
    sleep 5
    #monitor_pid=$(od -d $STATUS_FILE | head -n 2 | tail -n 1 | awk '{print $2}')
    monitor_pid=$(ps axo pid,cmd | grep "${APP_NAME}/bin/${APP_NAME}" | grep -vP "supervise|grep" | head -n 2 | tail -n 1 | awk '{print $1}')
    if [ -d /proc/$monitor_pid ]
    then
        _message "$ACTS start succ: pid=$monitor_pid"
    else
        _warning "$ACTS start failed"
        exit 1
    fi

}

foreground_start() {
    # 根据环境变量解析真实端口
    real_port=${PORT_MAIN}
    idc=${IDC_MAIN}
    # 设置目标conf位置
    service_conf=${ACTS_DIR}/conf/app_template.toml

    # 将template配置中的端口、idc替换生成正式conf
    sed -e "s@MATRIX_PORT_MAIN@${real_port}@g;s@IDC@${idc}@g;" conf/app_template.toml > app.toml
    # 前台启动程序
    $CMD
}


control_stop() {
    _killgroup
    sleep 5
}

control_restart() {
    control_stop
    control_start
}

control_check() {
    if is_running
    then
        monitor_pid=$(od -d $STATUS_FILE | head -n 2 | tail -n 1 | awk '{print $2}')
        _message "$ACTS is running: pid=$monitor_pid"
    else
        _message "$ACTS is not running"
    fi
}

control_check_log() {
    tail -7 $STATUS_DIR/supervise.log*
}

control_help() {
    _message "Usage: $(basename "$0") [start|stop|restart|check|check_log]"
}

check_with_case() {
    res=$(python ${ACTS_DIR}/../test_tools/bin/test_getfbx.py)
    if [ $res -eq 1 ]
    then
        _message "request server not exists"
    fi
    _message "check case success"
}

ACTION=$1

cd $ACTS_DIR

case "X$ACTION" in
    Xforeground_start)
        Xforeground_start
        ;;
    Xcheck_with_case)
        check_with_case
        ;;
    Xstart)
        control_start
        ;;
    Xrestart)
        control_restart
        ;;
    Xstop)
        control_stop
        ;;
    Xcheck)
        control_check
        ;;
    Xcheck_log)
        control_check_log
        ;;
    *)
        control_help
        ;;
esac
