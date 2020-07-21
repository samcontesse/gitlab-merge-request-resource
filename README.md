# GitLab Merge Request Concourse Resource

A concourse resource to check for new merge requests on GitLab and update the merge request status.

## Source Configuration

```yaml
resource_types:
- name: merge-request
  type: docker-image
  source:
    repository: samcontesse/gitlab-merge-request-resource

resources:
- name: merge-request
  type: merge-request
  source:
    uri: https://gitlab.com/myname/myproject.git
    private_token: XXX
```

* `uri`: (required) The location of the repository (required)
* `private_token`: (required) Your GitLab user's private token (required, can be found in your profile settings)
* `insecure`: When set to `true`, SSL verification is turned off 
* `skip_work_in_progress`: When set to `true`, merge requests mark as work in progress (WIP) will be skipped. Default `false`
* `skip_not_mergeable`: When set to `true`, merge requests not marked as mergeable will be skipped. Default `false`
* `skip_trigger_comment`: When set to `true`, the resource will not look up for `[trigger ci]` merge request comments to manually trigger builds. Default `false`  
* `concourse_url`: When set, this url will be used to override `ATC_EXTERNAL_URL` during commit status updates. No set default.  
* `pipeline_name`(string): When set, this url will be used to override `BUILD_PIPELINE_NAME` during commit status updates.  
* `labels`(string[]): Filter merge requests by label`[]`
* `target_branch`(string): Filter merge requests by target_branch. Default is empty string.

## Behavior

### `check`: Check for new merge requests

Checks if there are new merge requests or merge requests with new commits.

### `in`: Clone merge request source branch

`git clone`s the source branch of the respective merge request.

### Parameters

* `submodules`: Optional. If `none`, submodules will not be fetched. If not specified, or if `all` is explicitly specified, all submodules are fetched.

### `out`: Update a merge request's merge status

Updates the merge request's `merge_status` which displays nicely in the GitLab UI and allows to only merge changes if they pass the test.

#### Parameters

* `repository`: The path of the repository of the merge request's source branch (required)
* `status`: The new status of the merge request (required, can be either `pending`, `running`, `success`, `failed`, or `canceled`)
* `labels`(string[]): The labels you want add to your merge request
* `comment`: Add a comment for MR. Could be an object with `text`/`file` fields. If just the `file` or `text` is specified it is used to populate the field, if both `file` and `text` are specified then the file is substituted in to replace $FILE_CONTENT in the text.

## Example

```yaml
jobs:
- name: sample-merge-request
  plan:
  - get: merge-request
    trigger: true
  - put: merge-request
    params:
      repository: merge-request
      status: running
  - task: unit-test
    file: merge-request/ci/tasks/unit-test.yml
  on_failure:
    put: merge-request
    params:
      repository: merge-request
      status: failed
  on_success:
    put: merge-request
    params:
      repository: merge-request
      status: success
      labels: ['unit-test', 'stage']
      comment:
        file: out/commt.txt
        text: |
          Add new comment.
          $FILE_CONTENT
```
