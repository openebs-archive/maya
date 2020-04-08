#!/bin/bash
set -e

if [ "$#" -ne 3 ]; then
    echo "Error: Unable to create a new release. Missing required input."
    echo "Usage: $0 <github org/repo> <tag-name> <branch-name>"
    echo "Example: $0 kmova/bootstrap v1.0.0 master"
    exit 1
fi

C_GIT_URL=$(echo "https://api.github.com/repos/$1/releases")
C_GIT_TAG_NAME=$2
C_GIT_TAG_BRANCH=$3

if [ -z ${GIT_NAME} ];
then
  echo "Error: Environment variable GIT_NAME not found. Please set it to proceed.";
  echo "GIT_NAME should be a valid GitHub username.";
  exit 1
fi

if [ -z ${GIT_TOKEN} ];
then
  echo "Error: Environment variable GIT_TOKEN not found. Please set it to proceed.";
  echo "GIT_TOKEN should be a valid GitHub token associated with GitHub username.";
  echo "GIT_TOKEN should be configured with required permissions to create new release.";
  exit 1
fi

RELEASE_CREATE_JSON=$(echo \
{ \
 \"tag_name\":\"${C_GIT_TAG_NAME}\", \
 \"target_commitish\":\"${C_GIT_TAG_BRANCH}\", \
 \"name\":\"${C_GIT_TAG_NAME}\", \
 \"body\":\"Release created via $0\", \
 \"draft\":false, \
 \"prerelease\":false \
} \
)

#delete the temporary response file that might 
#have been left around by previous run of the command
#using a fixed name means that this script 
#is not thread safe. only one execution is permitted 
#at a time.
TEMP_RESP_FILE=temp-curl-response.txt
rm -rf ${TEMP_RESP_FILE}

response_code=$(curl -u ${GIT_NAME}:${GIT_TOKEN} \
 -w "%{http_code}" \
 --silent \
 --output ${TEMP_RESP_FILE} \
 --url ${C_GIT_URL} \
 --request POST --header 'content-type: application/json' \
 --data "$RELEASE_CREATE_JSON")

#When embedding this script in other scripts like travis, 
#success responses like 200 can mean error. rc_code maps
#the responses to either success (0) or error (1)
rc_code=0

#Github returns 201 Created on successfully creating a new release
#201 means the request has been fulfilled and has resulted in one 
#or more new resources being created.
if [ $response_code != "201" ]; then
    echo "Error: Unable to create release. See below response for more details"
    #The GitHub error response is pretty well formatted.
    #Printing the body gives all the details to fix the errors
    #Sample response when the branch already exists looks like this:
    #{
    #  "message": "Validation Failed",
    #  "errors": [
    #    {
    #      "resource": "Release",
    #      "code": "already_exists",
    #      "field": "tag_name"
    #    }
    #  ],
    #  "documentation_url": "https://developer.github.com/v3/repos/releases/#create-a-release"
    #}
    rc_code=1
else
    #Note. In case of success, lots of details of returned, but just 
    #knowing that creation worked is all that matters now.
    echo "Successfully tagged $1 with release tag ${C_GIT_TAG_NAME} on branch ${C_GIT_TAG_BRANCH}"
fi
cat ${TEMP_RESP_FILE}

#delete the temporary response file
rm -rf ${TEMP_RESP_FILE}

exit ${rc_code}
