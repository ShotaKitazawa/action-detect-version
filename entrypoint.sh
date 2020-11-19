#!/bin/bash

set -e

repos_url=${INPUT_REPOS_URL}
pr_url=${INPUT_PR_URL}
version_dir=${INPUT_DIR}

contents=$(curl -sL \
  -H "Accept: application/vnd.github.v3+json" \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  $pr_url/files)
filenames=$(echo $contents | jq '[.[].filename]')
len=$(echo $filenames | jq length)

detected_version=""
for i in $(seq 0 $(($len-1))); do
  filename=$(echo $filenames | jq -r .[$i])
  if [[ !($filename =~ ^$version_dir) ]]; then continue; fi
  v=$(echo $filename | sed -e "s|^$version_dir||")
  if [ "$detected_version" == "" ]; then detected_version=${v%%/*}; continue; fi
  if [ "$detected_version" != "${v%%/*}" ]; then echo "error: updated multiple version"; exit 1; fi
done
if [ "$detected_version" == "" ]; then echo "error: nothing updated"; exit 1; fi

echo "::set-output name=new_version::${detected_version}"
