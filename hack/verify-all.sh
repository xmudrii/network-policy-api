#!/bin/bash

# Copyright 2022 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${SCRIPT_ROOT}/hack/kube-env.sh"

SILENT=true

function is-excluded {
  for e in $EXCLUDE; do
    if [[ $1 -ef ${BASH_SOURCE} ]]; then
      return
    fi
    if [[ $1 -ef "$SCRIPT_ROOT/hack/$e" ]]; then
      return
    fi
  done
  return 1
}

while getopts ":v" opt; do
  case $opt in
    v)
      SILENT=false
      ;;
    \?)
      echo "Invalid flag: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

if $SILENT ; then
  echo "Running in the silent mode, run with -v if you want to see script logs."
fi

EXCLUDE="verify-all.sh"

ret=0
for t in `ls $SCRIPT_ROOT/hack/verify-*.sh`
do
  if is-excluded $t ; then
    echo "Skipping $t"
    continue
  fi
  if $SILENT ; then
    echo -e "Verifying $t"
    if bash "$t" &> /dev/null; then
      echo -e "${color_green}SUCCESS${color_norm}"
    else
      echo -e "${color_red}FAILED${color_norm}"
      ret=1
    fi
  else
    if bash "$t"; then
      echo -e "${color_green}SUCCESS: $t ${color_norm}"
    else
      echo -e "${color_red}Test FAILED: $t ${color_norm}"
      ret=1
    fi
  fi
done
exit $ret
