## Gitreport

Small tool for generating report of changes performed in Git repository.

Report is fully reproducible and all query parameters are stored in the filename for convenience.

## Usage

### Installation

To install, simply run:

```bash
$ go install github.com/wojciechka/gitreport/cmd/gitreport
```

### Usage

To perform a report of changes since specific time, do:

```bash
$ gitreport -output /tmp -from "2019-11-01" ~/go/src/github.com/wojciechka/gitreport
```

This will output report as `/tmp/gitreport-LQUTEKRlQweDVkYmI3NTgwClRUMHg1ZGM4NDA1Mw.txt` or similar filename.

To recreate the same report, simply run the following command:

```bash
gitreport -query gitreport-LQUTEKRlQweDVkYmI3NTgwClRUMHg1ZGM4NDA1Mw.txt ~/go/src/github.com/wojciechka/gitreport
```

This will a report, parsing the `LQUTEKRlQweDVkYmI3NTgwClRUMHg1ZGM4NDA1Mw` query into time range as well as any other parameters for the query.

### Options

The following options are accepted by the `gitreport` command:


|Option|Default value|Description|
| ---- | ----------- | --------- |
|`-output`|.|Output directory to write to; defaults to current working directory|
|`-ref`|*(unset)*|Tag or commit to generate history for; if not set, starts from current `HEAD`|
|`-all`|false|Rather than following a single ref, generate report for all commits in the repository|
|`-author`|*(unset)*|Regular expression of author to match commits for; useful for generating reports from specific person or people|
|`-from`|*(unset)*|Start date and/or time from which commits should be included in report; if not set, all commits before `-to` are included|
|`-since`|*(unset)*|Duration from now after which commits should be included in report; if not set, all commits before `-to` are included; can't be specified together with `-from`|
|`-to`|*(unset)*|End date and/or time to which commits should be included in report; if not set, all commits after `-from` / `-since` are included|
|`-query`|*(unset)*|Instead of using `-author`, `-from`, `-since` and `-to`, parse query from filename and re-generate same report|
