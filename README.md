# The Jira towel

[![Build Status](https://api.cirrus-ci.com/github/marco-m/jira-towel.svg?branch=master)](https://cirrus-ci.com/github/marco-m/jira-towel)

_You can wrap it around you for warmth as you bound across the cold moons of Jaglan Beta._

## Status

Very early stage. Unstable. API breakage will happen.

## Contributing and Development

This document explains how to use the tool. See [CONTRIBUTING](./CONTRIBUTING.md) for how to develop, test and contribute.

Please, before opening a PR, open a ticket to discuss your use case. This allows to better understand the why of a new feature and not to waste your time (and ours) developing a feature that for some reason doesn't fit well with the spirit of the project or could be implemented differently. This is in the spirit of [Talk, then code](https://dave.cheney.net/2019/02/18/talk-then-code).

We care about code quality, readability and tests, so please follow the current style and provide adequate test coverage. In case of doubts about how to tackle testing something, feel free to ask.

## Credentials

https://id.atlassian.com/manage-profile/security/api-tokens

## JQL

list all issues in a given epic
`parentEpic = MANGO-1`
where MANGO-1 is the key of the epic, while the epic name is "Odissea" :-/
but if the project is a TMP project (whatever that means) then the query is
`parent = MANGO-1`

https://community.atlassian.com/t5/Jira-questions/JQL-List-all-issues-of-Epic/qaq-p/1779069
https://community.atlassian.com/t5/Jira-questions/Query-to-find-all-issues-related-to-an-EPIC-including-subtasks/qaq-p/871457


list all issues in a list of epics
`parentEpic in (MANGO-1, MANGO-7)`

list all epics
`issuetype = Epic`
