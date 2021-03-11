package main

import (
	"flag"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	secChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/cmd/make/build"
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
		build.Publish()
	} else {
		l = logger.DefaultSLogger("build")
		build.Compile()
	}
}

func applyFlags() {

	build.BinName = secChecker.BinName
	build.OSSPath = secChecker.OSSPath

	build.BuildDir = *flagBuildDir
	build.PubDir = *flagPubDir
	build.Archs = *flagArchs

	build.Release = *flagEnv
	build.MainEntry = *flagMain
	build.DownloadAddr = *flagDownloadAddr
}
