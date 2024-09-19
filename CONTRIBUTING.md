# Contributing and developing jira-towel

Please, before opening a PR, open a ticket to discuss your use case. This allows to better understand the why of a new feature and not to waste your time (and ours) developing a feature that for some reason doesn't fit well with the spirit of the project or could be implemented differently. This is in the spirit of [Talk, then code](https://dave.cheney.net/2019/02/18/talk-then-code).

We care about code quality, readability and tests, so please follow the current style and provide adequate test coverage. In case of doubts about how to tackle testing something, feel free to ask.

## Credentials

https://id.atlassian.com/manage-profile/security/api-tokens



## Development

Jira HTTP documentation: https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#about

Basic auth for REST API
https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/

CAPTCHA
https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/#captcha

> A CAPTCHA is 'triggered' after several consecutive failed log in attempts. If CAPTCHA has been triggered, you cannot use Jira's REST API to authenticate.
>
> You can check this in the error response from Jira. If there is an `X-Seraph-LoginReason` header with a value of `AUTHENTICATION_DENIED`, the application rejected the login without even checking the password. This is the most common indication that Jira's CAPTCHA feature has been triggered.

Status codes and detailed errors
https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#status-codes

> Operations that return an error status code may also return a response body containing details of the error or errors. The schema for the response body is shown below:

```json
{
  "id": "https://docs.atlassian.com/jira/REST/schema/error-collection#",
  "title": "Error Collection",
  "type": "object",
  "properties": {
    "errorMessages": {
      "type": "array",
      "items": {
        "type": "string"
      }
    },
    "errors": {
      "type": "object",
      "patternProperties": {
        ".+": {
          "type": "string"
        }
      },
      "additionalProperties": false
    },
    "status": {
      "type": "integer"
    }
  },
  "additionalProperties": false
}
```

Expansion, pagination, and ordering:
https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#expansion

Expansion
> The Jira REST API uses resource expansion, which means that some parts of a resource are not returned unless specified in the request. This simplifies responses and minimizes network traffic.

Pagination
> The Jira REST API uses pagination to improve performance. Pagination is enforced for operations that could return a large collection of items. When you make a request to a paginated resource, the response wraps the returned array of values in a JSON object with paging metadata.
>
> Each operation can have a different limit for the number of items returned, and these limits may change without notice.

The text also explains how to adapt to any limit.

Ordering
> Some operations support ordering the elements of a response by a field.

## JQL

> if you don't specify any JQL parameter here, it's the same as if you searched in the issue navigator in Jira and left the advanced search field empty; it returns all the issues that this user has access to.

(modulo the paginating behavior)
