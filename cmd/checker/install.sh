#! /bin/bash

download_url="https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/security-checker-"

goos="$(uname -s)"
goarch="$(uname -m)"

if [ "${goos}" == "Linux" ] || [ "${goos}" == "linux" ]  ;then
    goos="linux"
fi

if [ "${goarch}" == "x86_64" ] ;then
    goarch="amd64"
fi

download_url="${download_url}${goos}-${goarch}.tar.gz"


service_name="security-checker"
bin_name="checker"
install_dir="/usr/local/${service_name}"
upgrade=${1}

if ! mkdir -p "${install_dir}" ;then
    exit -1
fi

while [[ $# > 1 ]]
do
  opt="$1"

  case $opt in

    --upgrade)
      upgrade=1
      shift
      ;;
      
    *)
      # unknown option
    ;;
  esac
  shift
done


echo "stopping ${service_name} ..."

systemctl stop ${service_name} &>/dev/null
service ${service_name} &>/dev/null
stop ${service_name} &>/dev/null

echo "downloading..."
wget -O - "${download_url}" | tar -xz -C "${install_dir}"
if [ $? -ne 0 ] ;then
    exit -1
fi

init() {
    mkdir -p "${install_dir}/rules.d"
    if [ ! -f "${install_dir}/checker.conf" ] ;then
        mv "${install_dir}/checker.conf.sample"  "${install_dir}/checker.conf"
    fi
}

registerService() {
    if hash systemctl 2>/dev/null; then

    systemctl stop ${service_name}.service > /dev/null 2>&1
    systemctl disable ${service_name}.service > /dev/null 2>&1

    systemd_path="/lib/systemd"
    if [ ! -e /lib/systemd ]; then
        # suse/aliyun-linux 是在 /etc/systemd 下, 包括 opensuse 以及其他 linux, 都在 /lib/systemd 下
        systemd_path="/etc/systemd" 
    fi

    # 有些系统中, 系统目录(/lib/systemd, /usr/local/ 是只读的, 比如 coreos)
    if cp ${install_dir}/autostart/${bin_name}.service ${systemd_path}/system/${service_name}.service ; then
        systemctl enable ${service_name}.service
        systemctl start ${service_name}.service
    else
        echo "failed"
        exit -1
    fi
    else
    if hash init-checkconf 2>/dev/null; then

        service ${service_name} stop > /dev/null 2>&1

        if cp ${install_dir}/autostart/${bin_name}.conf /etc/init/${service_name}.conf ; then
            service ${service_name} start
        else
            echo "failed"
            exit -1
        fi
    else
        if hash initctl 2>/dev/null; then
        service ${service_name} stop > /dev/null 2>&1
        stop ${service_name} > /dev/null 2>&1

        if cp ${install_dir}/autostart/${bin_name}.conf /etc/init/${service_name}.conf ; then
            start ${service_name}
        else
            echo "failed"
            exit -1
        fi
        else
        
        service ${service_name} stop > /dev/null 2>&1

        if cp ${install_dir}/autostart/${bin_name}.sh /etc/init.d/${service_name} ; then
            chmod +x /etc/init.d/${service_name}

            if hash update-rc.d 2>/dev/null; then
            update-rc.d ${service_name} defaults > /dev/null 2>&1
            else
            chkconfig --add ${service_name} > /dev/null 2>&1
            fi

            service ${service_name} start
        else
            echo "failed"
            exit -1
        fi

        fi
    fi
    fi
} 

init

echo "register service..."
registerService

echo "Success!"

