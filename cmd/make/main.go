package main

import (
	"flag"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	secChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
)

var (
	flagMain         = flag.String(`main`, `main.go`, `binary build entry`)
	flagBuildDir     = flag.String("build-dir", "build", "output of build files")
	flagDownloadAddr = flag.String("download-addr", "", "")
	flagPubDir       = flag.String("pub-dir", "pub", "")
	flagArchs        = flag.String("archs", "local", "os archs")
	flagEnv          = flag.String(`env`, ``, `build for local/test/preprod/release`)
	flagPub          = flag.Bool(`pub`, false, `publish binaries to OSS: test/release`)

	binName string
	ossPath string

	l *logger.Logger
)

func main() {
	flag.Parse()
	applyFlags()

	if *flagPub {
		l = logger.DefaultSLogger("pub")
		publish()
	} else {
		l = logger.DefaultSLogger("build")
		compile()
	}
}

func applyFlags() {

	binName = secChecker.BinName
	ossPath = secChecker.OSSPath

	BuildDir = *flagBuildDir
	PubDir = *flagPubDir
	Archs = *flagArchs

	Release = *flagEnv
	MainEntry = *flagMain
	DownloadAddr = *flagDownloadAddr
}
