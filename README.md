# Action Detect Version

This is a GitHub Action to detect version name from updated directory name in Pull Request.

For example:

* case 1
    * inputs: `dir=path/to/dir/`
    * updated files:
        * `path/to/dir/v0.0.1/hoge`
        * `path/to/dir/v0.0.1/fuga`
        * `path/to/dir/v0.0.1/piyo`
    * then, outpus: `new_version=v0.0.1`
* case 2
    * inputs: `dir=path/to/dir/`
    * updated files:
        * `path/hoge`
    * then, Action is error occurs (msg: `error: nothing updated`)
* case 3
    * inputs: `dir=path/to/dir/`
    * updated files:
        * `path/to/dir/v0.0.1/hoge`
        * `path/to/dir/v0.0.1/fuga`
        * `path/to/dir/v0.0.2/hoge`
    * then, Action is error occurs (msg: `error: updated multiple version`)

## Inputs

| NAME     | DESCRIPTION                                                              | TYPE     |
|----------|--------------------------------------------------------------------------|----------|
| `dir`    | specify base directory                                                   | `string` |
| `pr_url` | target PullRequest URL (fixed as `${{ github.event.pull_request.url }}`) | `string` |

## Outputs

| NAME          | DESCRIPTION             | TYPE     |
|---------------|-------------------------|----------|
| `new_version` | updated directory name. | `string` |

## Example

* if PullRequest labeled `update-template`, run bellow processes.
    1. `Detect template version`: get directory name that updated files under the `templates/`
    2. `Get sentence of Release Note from Pull Request`: get PullRequest body using `actions-ecosystem/action-regex-match@v2`
    3. `Push Tag`: push tag of the directory name output on process 1
    4. `Create Release`: create release from tag pushed process 3 & body output of process 2

```
name: Push a new tag & release with merged Pull Request

on:
  pull_request:
    types: [closed]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Detect template version
        uses: ShotaKitazawa/action-detect-version@main
        id: detect-version
        if: ${{ github.event.pull_request.merged == true && contains(github.event.pull_request.labels.*.name, 'update-template') }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          pr_url: ${{ github.event.pull_request.url }}
          dir: templates/

      - name: Get sentence of Release Note from Pull Request
        uses: actions-ecosystem/action-regex-match@v2
        id: regex-match
        if: ${{ steps.detect-version.outputs.new_version != null }}
        with:
          text: ${{ github.event.pull_request.body }}
          regex: '```release_note([\s\S]*)```'

      - name: Push Tag
        uses: actions-ecosystem/action-push-tag@v1
        id: push-tag
        if: ${{ steps.detect-version.outputs.new_version != null && steps.regex-match.outputs.match != '' }}
        with:
          tag: ${{ steps.detect-version.outputs.new_version }}

      - name: Create Release
        uses: actions/create-release@v1
        id: create-release
        if: ${{ steps.detect-version.outputs.new_version != null && steps.regex-match.outputs.match != '' }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ steps.detect-version.outputs.new_version }}
          release_name: ${{ steps.detect-version.outputs.new_version }}
          body: ${{ steps.regex-match.outputs.group1 }}
          draft: false
          prerelease: false
```

