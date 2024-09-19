# The Jira towel

[![Build Status](https://api.cirrus-ci.com/github/marco-m/jira-towel.svg?branch=master)](https://cirrus-ci.com/github/marco-m/jira-towel)

_You can wrap it around you for warmth as you bound across the cold moons of Jaglan Beta_ ([Towel Day](https://en.wikipedia.org/wiki/Towel_Day)).

## Status

Very early stage. Unstable. API breakage will happen.

## Contributing and Development

This document explains how to use the tool. See [CONTRIBUTING](./CONTRIBUTING.md) for how to develop, test and contribute.

**Please, before opening a PR, open a ticket to discuss your use case**. This allows to better understand the why of a new feature and not to waste your time (and ours) developing a feature that for some reason doesn't fit well with the spirit of the project or could be implemented differently. This is in the spirit of [Talk, then code](https://dave.cheney.net/2019/02/18/talk-then-code).

We care about code quality, readability and tests, so please follow the current style and provide adequate test coverage. In case of doubts about how to tackle testing something, feel free to ask.

## Credentials

- Run `jira-towel init`.
- Go to the [API tokens](https://id.atlassian.com/manage-profile/security/api-tokens) page of your Atlassian account and create an API token.
- Store it in the `jira-towel.json` configuration file, created in the previous step with `jira-towel init`. Ensure that its permissions are `0600` (only the owner can read it).

## Surviving Jira custom fields

Jira has a feature that is useful from the point of view of the user, but with a annoying implementation for a consumer of the API, "custom fields".

For example, if project `BANANA` has a custom field `My Product`, then it can be easily referenced in a query:

```
project = BANANA AND "My Product" = "Continuous Deployment"
```

The problem is the encoding of such information in the API reply:

```json
{
  "fields": {
    "customfield_12011": {
      "value": "Continuous Deployment"
      ...
    }
  }
}
```

So a program has to beforehand discover how to map name `My Product` to the number `12011` and finally construct the string `customfield_12011` and replace it on the fly with the name `My Product` everywhere in the JSON object from the reply.

Once you have found the numeric ID of the custom field you are interested in, you can specify it as follows:

```
jira-towel graph --custom-fields=product:11919 graph --cluster-by=product
```

## Creating the graph

Command `jira-towel graph` generates the graphviz file `graph.dot`. You can render it for example to SVG:

```
dot -Tsvg graph.dot > planned.svg
open -a Firefox planned.svg
```

## JQL Examples

API documentation for [JQL search](https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issue-search/#api-rest-api-2-search-post).

First, you need to specify the project, otherwise it will query ALL the projects available to your account: `--jql 'project = <PROJECT NAME> AND ...'`

```
'project=<PROJECT> AND issuetype = Epic and status in ("to do", "in progress")'
```

All epics in a given custom field, any state:

```
project = <PROJECT> AND issuetype = Epic AND "Product[Dropdown]" = "My Product"
```

List all issues in a given epic

```
parentEpic = MANGO-1
```

Where MANGO-1 is the key of the epic, while the epic name is "Odissea" :-/
but if the project is a TMP project (whatever that means) then the query is
`parent = MANGO-1`

https://community.atlassian.com/t5/Jira-questions/JQL-List-all-issues-of-Epic/qaq-p/1779069
https://community.atlassian.com/t5/Jira-questions/Query-to-find-all-issues-related-to-an-EPIC-including-subtasks/qaq-p/871457


List all issues in a list of epics
```
parentEpic in (MANGO-1, MANGO-7)
```

List all epics
```
issuetype = Epic
```
