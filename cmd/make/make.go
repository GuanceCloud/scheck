package main

import (
	"flag"
	"os"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/cmd/make/build"
)

var (
	flagBinary       = flag.String("binary", "", "binary name to build")
	flagBuildDir     = flag.String("build-dir", "build", "output of build files")
	flagMain         = flag.String(`main`, `main.go`, `binary build entry`)
	flagDownloadAddr = flag.String("download-addr", "", "")
	flagPubDir       = flag.String("pub-dir", "pub", "")
	flagArchs        = flag.String("archs", "local", "os archs")
	flagEnv          = flag.String(`env`, ``, `build for local/test/preprod/release`)
	flagPub          = flag.Bool(`pub`, false, `publish binaries to OSS: local/test/release/preprod`)
)

func applyFlags() {

	build.AppBin = *flagBinary
	build.BuildDir = *flagBuildDir
	build.PubDir = *flagPubDir
	build.Archs = *flagArchs

	build.Release = *flagEnv
	build.ReleaseType = build.Release
	build.MainEntry = *flagMain
	build.DownloadAddr = *flagDownloadAddr

	if *flagPub {
		build.PubDatakit()
		os.Exit(0)
	}
}

func main() {
	flag.Parse()
	applyFlags()
	build.Compile()
}
