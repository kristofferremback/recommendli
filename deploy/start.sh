#!/usr/bin/env bash

script_dir=$(realpath $(dirname "${BASH_SOURCE[0]}"))
executable_path=$(make executable-path)

if [[ ! -f "${executable_path}" ]]; then
	curr_pwd=$(pwd)
	cd "${script_dir}/.."

	make build

	cd "${curr_pwd}"
fi

"${executable_path}" ${@}
