#/bin/bash
# date: Wed Jul 21 11:29:22 CST 2021
# author: tan

branch_name="$(git symbolic-ref HEAD 2>/dev/null)" ||
branch_name="(unnamed branch)"     # detached HEAD

branch_name=${branch_name##refs/heads/}

new_tag=$1
latest_tag=$(git describe --abbrev=0 --tags)

case $branch_name in
	"testing")
		# TODO
		;;

	"master") echo "release prod release..."
		if [ -z $new_tag ]; then
			echo "[E] new tag required to release production datakit, latest tag is ${latest_tag}"
		else
			git tag -f $new_tag  &&

			# Trigger CI to release other platforms
			git push -f --tags   &&
			git push
		fi
		;;

	"github") echo "release to github"
		# TODO
		;;

	*) echo "[E] unsupported branch '$branch_name' for release, exited"
		;;
esac
