package impl

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/container/impl/utils"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

var (
	optionSplitChar = global.OptionSplitChar[global.LocalGOOS]
	log             = logger.DefaultSLogger("container")
)

func Inspect(facts map[string]interface{}, pretty bool) string {
	imageName, ok := utils.GetValueN(facts, "Image").(string)
	if !ok {
		log.Warnf("utils.GetValueN is not string")
		return ""
	}
	image, err := utils.ContainerInspectMap(imageName)
	if err != nil {
		log.Warnf("err=%v", err)
		return ""
	}

	stringBuilder := strings.Builder{}
	stringBuilder.WriteString("docker run")

	// -d
	stringBuilder.WriteString(boolOption(facts, "--detach", false, "Config", "AttachStdout"))
	// -ti
	stringBuilder.WriteString(boolOption(facts, "--tty", true, "Config", "Tty"))
	stringBuilder.WriteString(boolOption(facts, "--interactive", true, "Config", "AttachStdin"))
	// --rm
	stringBuilder.WriteString(boolOption(facts, "--rm", true, "HostConfig", "AutoRemove"))
	// --restart
	stringBuilder.WriteString(parseRestart(facts))
	// --name
	stringBuilder.WriteString(parseName(facts))
	// --pid
	stringBuilder.WriteString(stringOption(facts, "--pid", "", "HostConfig", "PidMode"))
	// --ipc
	stringBuilder.WriteString(stringOption(facts, "--ipc", "private", "HostConfig", "IpcMode"))
	// --network
	stringBuilder.WriteString(stringOption(facts, "--network", "default", "HostConfig", "NetworkMode"))
	// --privileged
	stringBuilder.WriteString(boolOption(facts, "--privileged", true, "HostConfig", "Privileged"))
	// -p, --port
	stringBuilder.WriteString(parseBindingPort(facts))
	// --hostname
	stringBuilder.WriteString(parseHostname(facts))
	// --user
	stringBuilder.WriteString(stringOption(facts, "--user", "", "Config", "User"))
	// --mac-address
	stringBuilder.WriteString(stringOption(facts, "--mac-address", "", "Config", "MacAddress"))
	// --link
	stringBuilder.WriteString(parseLinks(facts))
	// --add-host
	stringBuilder.WriteString(stringArrayOptions(facts, "--add-host", "HostConfig", "ExtraHosts"))
	// -e, --env
	stringBuilder.WriteString(parseEnv(facts, image))
	// -v, --volume
	stringBuilder.WriteString(stringArrayOptions(facts, "-v", "HostConfig", "Binds"))
	stringBuilder.WriteString(stringArrayOptions(facts, "-v", "Config", "Volumes"))
	// --volumes-from
	stringBuilder.WriteString(stringArrayOptions(facts, "--volumes-from", "HostConfig", "VolumesFrom"))
	// --cap-add
	stringBuilder.WriteString(stringArrayOptions(facts, "--cap-add", "HostConfig", "CapAdd"))
	// --cap-drop
	stringBuilder.WriteString(stringArrayOptions(facts, "--cap-drop", "HostConfig", "CapDrop"))
	// --device
	stringBuilder.WriteString(parseDevices(facts))
	// --dns
	stringBuilder.WriteString(stringArrayOptions(facts, "--dns", "HostConfig", "Dns"))
	// --log-driver, --log-opt
	stringBuilder.WriteString(parseLog(facts))
	// --label
	stringBuilder.WriteString(parseLabels(facts, image))
	// --entrypoint
	entry, args := parseEntrypoint(facts, image)
	stringBuilder.WriteString(entry)
	// 镜像URL
	stringBuilder.WriteString(optionSplitChar)
	stringBuilder.WriteString(utils.GetValueN(facts, "Config", "Image").(string))
	// Args
	if cmd := parseCmd(facts); cmd != "" {
		args += cmd
	}
	args = strings.ReplaceAll(args, optionSplitChar, " ")
	if args != "" {
		stringBuilder.WriteString(optionSplitChar)
		stringBuilder.WriteString(strings.TrimSpace(args))
	}
	return stringBuilder.String()
}

func parseBindingPort(facts map[string]interface{}) string {
	ports, ok := utils.GetValueN(facts, "NetworkSettings", "Ports").(map[string]interface{})
	if !ok || ports == nil {
		return ""
	}
	stringBuilder := strings.Builder{}
	for cPort, hostPorts := range ports {
		splits := strings.Split(cPort, "/")
		containerPort := splits[0]
		protocol := splits[1]
		if hostPorts != nil {
			for _, m := range hostPorts.([]interface{}) {
				hPort, ok := m.(map[string]interface{})
				if !ok {
					continue
				}
				hostIP := hPort["HostIp"]
				hostPort := hPort["HostPort"]

				if hostIP != "" && hostIP != "0.0.0.0" {
					stringBuilder.WriteString(fmt.Sprintf("%s-p %s:%s:%s",
						optionSplitChar, hostIP, hostPort, containerPort))
				} else {
					stringBuilder.WriteString(fmt.Sprintf("%s-p %s:%s",
						optionSplitChar, hostPort, containerPort))
				}
				if protocol != "tcp" {
					stringBuilder.WriteString(fmt.Sprintf("/%s", protocol))
				}
			}
		}
	}
	return stringBuilder.String()
}

func parseHostname(facts map[string]interface{}) string {
	hostname := utils.GetValueN(facts, "Config", "Hostname").(string) // nolint
	return fmt.Sprintf("%s--hostname %s", optionSplitChar, hostname)
}

func stringOption(facts map[string]interface{}, optName, def string, keys ...string) string {
	optValue := utils.GetValueN(facts, keys...)
	optValueText := fmt.Sprintf("%v", optValue)
	if optValue != nil && optValueText != "" && optValueText != def {
		return fmt.Sprintf("%s%s %s", optionSplitChar, optName, optValueText)
	}
	return ""
}

func boolOption(facts map[string]interface{}, opt string, expectValue bool, keys ...string) string {
	res := utils.GetBool(facts, keys...)
	if res == expectValue {
		return fmt.Sprintf("%s%s", optionSplitChar, opt)
	}
	return ""
}

func stringArrayOptions(facts map[string]interface{}, optName string, keys ...string) string {
	rawValues, ok := utils.GetValueN(facts, keys...).([]interface{})
	if !ok || len(rawValues) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for _, rawVal := range rawValues {
		sb.WriteString(fmt.Sprintf("%s%s %v", optionSplitChar, optName, rawVal))
	}
	return sb.String()
}

func parseLinks(facts map[string]interface{}) string {
	links, ok := utils.GetValueN(facts, "HostConfig", "Links").([]interface{})
	if ok && len(links) > 0 {
		stringBuilder := strings.Builder{}
		for _, l := range links {
			link, ok := l.(string)
			if !ok {
				continue
			}
			splits := strings.Split(link, ":")
			s1 := strings.Split(splits[0], "/")
			src := s1[len(s1)-1]
			s2 := strings.Split(splits[1], "/")
			dst := s2[len(s2)-1]
			if src != dst {
				stringBuilder.WriteString(fmt.Sprintf("%s--link %s:%s", optionSplitChar, src, dst))
			} else {
				stringBuilder.WriteString(fmt.Sprintf("%s--link %s", optionSplitChar, src))
			}
		}
		return stringBuilder.String()
	}
	return ""
}

func parseRestart(facts map[string]interface{}) string {
	restart, ok := utils.GetValueN(facts, "HostConfig", "RestartPolicy", "Name").(string)
	if ok {
		if restart == "on-failure" {
			maxRetries, ok := utils.GetValueN(facts, "HostConfig", "RestartPolicy", "MaximumRetryCount").(int)
			if ok && maxRetries > 0 {
				max := fmt.Sprintf(":%d", maxRetries)
				return restart + max
			}
		} else if restart != "no" {
			return fmt.Sprintf("%s--restart=%s", optionSplitChar, restart)
		}
	}
	return ""
}

func parseEntrypoint(container, image map[string]interface{}) (entry, build string) {
	containerEntrypoint, ok := utils.GetValueN(container, "Config", "Entrypoint").([]interface{})
	if !ok || len(containerEntrypoint) == 0 {
		return "", ""
	}
	imageEntrypoint, ok := utils.GetValueN(image, "Config", "Entrypoint").([]interface{})
	if !ok {
		imageEntrypoint = []interface{}{}
	}
	// 如果容器Entrypoint与镜像Entrypoint相同，则不需要显示配置这个参数
	if fmt.Sprintf("%v", containerEntrypoint) == fmt.Sprintf("%v", imageEntrypoint) {
		return "", ""
	}

	entrypoint, ok := utils.GetValueN(container, "Config", "Entrypoint").([]interface{})
	if !ok || len(entrypoint) == 0 {
		return "", ""
	}
	args := strings.Builder{}
	for _, option := range entrypoint {
		if entry == "" {
			entry = fmt.Sprintf("%s--entrypoint=\"%s\"", optionSplitChar, option)
		} else {
			args.WriteString(optionSplitChar + option.(string))
		}
	}
	return entry, args.String()
}

func parseCmd(facts map[string]interface{}) string {
	cmd, ok := utils.GetValueN(facts, "Config", "Cmd").([]interface{})
	if !ok || len(cmd) == 0 {
		return ""
	}
	args := strings.Builder{}
	for _, option := range cmd {
		args.WriteString(optionSplitChar + option.(string))
	}
	return args.String()
}

func parseDevices(facts map[string]interface{}) string {
	devices, ok := utils.GetValueN(facts, "HostConfig", "Devices").([]interface{})
	if !ok || len(devices) == 0 {
		return ""
	}
	stringBuilder := strings.Builder{}
	for _, dc := range devices {
		deviceSpec := dc.(map[string]interface{})           // nolint
		host := deviceSpec["PathOnHost"].(string)           // nolint
		container := deviceSpec["PathInContainer"].(string) // nolint
		perms := deviceSpec["CgroupPermissions"].(string)   // nolint
		spec := fmt.Sprintf("%s:%s", host, container)
		if perms != "rwm" {
			spec += ":" + perms
		}
		stringBuilder.WriteString(fmt.Sprintf("%s--device %s", optionSplitChar, spec))
	}
	return stringBuilder.String()
}

func parseLabels(container, image map[string]interface{}) string {
	return parse(container, image, "Config", "Labels")
}

func parseLog(facts map[string]interface{}) string {
	stringBuilder := strings.Builder{}

	logType := utils.GetValueN(facts, "HostConfig", "LogConfig", "Type").(string) // nolint
	if logType != "json-file" {
		stringBuilder.WriteString(fmt.Sprintf("%s--log-driver=%s", optionSplitChar, logType))
	}

	logOpts, ok := utils.GetValueN(facts, "HostConfig", "LogConfig", "Config").(map[string]interface{})
	if ok && len(logOpts) > 0 {
		for key, val := range logOpts {
			stringBuilder.WriteString(fmt.Sprintf("%s--log-opt %s=%s", optionSplitChar, key, val))
		}
	}

	return stringBuilder.String()
}

func parseName(facts map[string]interface{}) string {
	name, ok := facts["Name"].(string)
	if ok {
		name = strings.Split(name, "/")[1]
		return fmt.Sprintf("%s--name %s", optionSplitChar, name)
	}
	return ""
}

func parseEnv(container, image map[string]interface{}) string {
	return parse(container, image, "Config", "Env")
}

func parse(container, image map[string]interface{}, config, env string) string {
	containerEnvs, ok := utils.GetValueN(container, config, env).([]interface{})
	if !ok || len(containerEnvs) == 0 {
		return ""
	}
	imageEnvs, ok := utils.GetValueN(image, config, env).([]interface{})
	if !ok {
		imageEnvs = []interface{}{}
	}
	// 只显示区别于镜像的环境变量
	res := strings.Builder{}
	for _, containerEnv := range containerEnvs {
		found := false
		for _, imageEnv := range imageEnvs {
			if containerEnv == imageEnv {
				found = true
				break
			}
		}
		if !found {
			res.WriteString(fmt.Sprintf("%s-e %v", optionSplitChar, containerEnv))
		}
	}
	return res.String()
}
