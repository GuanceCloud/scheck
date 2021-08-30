# Scheck developer's guide

## Development settings
### golang Environment configuration
```
# At present, our code is privately managed. Please tell the go compiler that this is a private repo warehouse address
go env -w GOPRIVATE=gitlab.jiagouyun.com/*
```
### git settings
When the code of our private repository is referenced in the project, the compiler will try to clone the code, but our code repository has user authentication. It is recommended to use SSH encryption free mode. Add the following configuration in ~ /. Gitconfig:
```
[url "ssh://git@gitlab.jiagouyun.com:40022/"]
	insteadOf = http://gitlab.jiagouyun.com/
```
At the same time, add SSH configuration (settings / SSH keys...) of your development machine in gitlab
### environment variables
Set the following environment variables and save them to `~/.scenv`
```
export LOCAL_OSS_ACCESS_KEY='LTAI5tLaYtUxxx'
export LOCAL_OSS_SECRET_KEY='nRr1xQBCeyl4oBgo0xx'
export LOCAL_OSS_BUCKET='df-xxx-dev'
export LOCAL_OSS_HOST='oss-cn-hangzhou.aliyuncs.com'
export SC_USERNAME='xxx'
```


## Build
### Early dependent installation
- packr2 install 
```
go get -u github.com/gobuffalo/packr/v2
go get -u github.com/gobuffalo/packr/v2/packr2
```
### Build local package
```
make local
```
### push local package
```
make pub_local
```
### install local package
```
# install.sh  In your project root directory
sh install.sh 
# upgrade Scheck
SC_UPGRADE=true ;sh install.sh
```
### other
```
# If you want to test whether multiple platforms are packaged successfully, you might as well try the following command
make local_all
make pub_local_all
```
