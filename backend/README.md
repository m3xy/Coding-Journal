# Table of Contents

1.  [Response Format](#response_format)
2.  [API](#api)
    1.  [Journal](#journal)
        1.  [Login](#journal_login)
    2.  [Authentication](#authentication)
        1.  [Login](#authentication_login)
        2.  [Register](#authentication_register)
    3.  [User](#user)
        1.  [Profile Query](#user_query)

<a id="response_format"></a>

# Response Format

Response bodies are always given in this below format (with the exception of journal endpoints) -

    interface Response {
    	message: string // Response's message - gives details on response status.
    	error: boolean  // Whether the request errored or not.
    	content: any    // The response's body.
    }

<a id="api"></a>

# API

This section describes main enpoints. There are 4 main entrypoints in the api,
starting with `cs3099user11.host.cs.st-andrews.ac.uk/api/<version>`:

- `/journal/`: Supergroup compliant endpoints,
- `/auth/`: Authentication requests - client login, signup.
- `/user/`: User retrieval requests.

<a id="journal"></a>

## Journal

<a id="journal_login"></a>

### Login

Supergroup compliant journal login.

1.  Endpoint

    This is accessed via the **GET** `/journal/login` endpoint.

2.  Request

    1.  Headers

        - X-FOREIGNJOURNAL-SECURITY-TOKEN: Journal&rsquo;s secret key. Needed for all journal queries. **MUST NOT BE SHARED WITH ANY CLIENT.**

    2.  Body

```json
{
  "email": "string",
  "password": "string"
}
```

3.  Response

    1.  Status

        - 200 - User logged in, response OK.
        - 404 - User was not found.
        - 401 - Access Unauthorized. See Headers.

    2.  Content

```typescript
interface Content {
  userId: string; // ID of the logged in user. Empty if errored
}
```

<a id="authentication"></a>

## Authentication

<a id="authentication_login"></a>

### Login

Client-only login endpoint.

1.  Endpoint

    This is accessed via the **GET** `/auth/login` endpoint.

2.  Request

    1.  Headers

        No headers required.

    2.  Body

```typescript
interface Body {
  client_id: string; // ID used to find client - here, it is an email.
  client_secret: string; // Client's secret for authentication - here, password.
}
```

3.  Response

    1.  Status

        - 200 - User logged in, response OK.
        - 404 - User was not found.
        - 401 - Access Unauthorized. See Headers.

    2.  Content

```typescript
interface Content {
  token: string; // Token used for restriced user access later on.
  refresh_token: string; // Refresh token - used to make a new token after expiry.
  redirect_uri: string; // Address to token refresh request.
}
```

<a id="authentication_register"></a>

### Register

Client registration endpoint.

1.  Endpoint

    The endpoint is accessible as **POST** `/auth/register`

2.  Request

    1.  Headers

        No required headers.

    2.  Body

```typescript
interface Content {
  // Required
  Email: string; // Client ID
  Password: string; // Client Secret
  FirstName: string;
  LastName: string;

  // Optional
  PhoneNumber?: string;
  Organization?: string;
}
```

3.  Response

    1.  Status

        - 200 - User successfully registered, response OK.
        - 405 - Bad request, form given is invalid.

    2.  Body

        This function has no response content.

<a id="user"></a>

## User

<a id="user_query"></a>

### Profile Query

User profile information query.

1.  Endpoint

    The endpoint for the query is **GET** `/user/{ID}`, where ID stands for the user&rsquo;s UUID in the server.

2.  Request

    1.  Headers

        No header or authentication is required for this query.

3.  Response

    1.  Status

        - 200 - Query successful, requested information contained in content.
        - 404 - User not found, content empty.

    2.  Body

        The \`Content\` value in the response is of type below

```typescript
interface Content {
  UserID: string;
  FullName: string;
  Profile: Profile;
}

interface Profile {
  Email: string;
  FirstName: string;
  LastName: string;
  PhoneNumber: string;
  Organization: string;
  CreatedAt: DateTime; // Format -
}
```

## Submission

### Get available tags

Queries the avaiable tags from the database to allow for their display on the frontend

1. Endpoint

    The endpoint is **GET** `/submissions/tags`

2. Request

    1. Headers
    2. Body

3. Response

    1. Status

        - 200 - if a non-empty tag array was found and returned
        - 204 - if no tags are currently in the db
        - 500 - if anything else goes wrong not pertaining to the client

    2. Body - the object shown below

```typescript
interface GetAvailableTagsResponse {
    message: string;
    error: bool;
    tags: string[];
}
```

### Submission Query

Querying an ordered list of submissions based upon query parameters

1. Enpoint

    The enpoint to query submissions is **GET** `/submissions/query`

2. Request

    1. Headers

    2. Body

    3. Parameters:

        - orderBy - newest or oldest
        - tags - any existant code tag (i.e. python, java, etc.)
        - authors - user ID of authors
        - reviewers - user ID of reviewers

3. Response

    1.  Status

        - 200 - if the submissions are queried properly
        - 204 - if the query returns an empty result set
        - 400 - if the query is malformatted (i.e. illegal query parameters)
        - 500 - if something else goes wrong in the backend

    2.  Body - the object shown below

```typescript 
interface QuerySubmissionsResponse {
    message: string;
    error: bool;
    submissions: Submission[];
}

interface Submission {
    id: uint;
    name: string;
}
```

### Submission Upload

Uploading submissions from the local journal

1. Enpoint

    The enpoint to upload submission is **POST** `/submissions/create`

2. Request

    1. Headers

    2. Body - the request body is shown below

```typescript
interface Submission {
    name: string;
    license: string;
    files: File[];
    authors: string[];
    reviewers: string[];
    categories: string[];
}

interface File {
    path: string;
    name: string;
    base64Value: string; // base64 encoded content
}

interface GlobalUser {
    userId: string
}
```

3. Response

    1.  Status

        - 200 - if the submision is uploaded properly
        - 400 - if the submission is not sent in the right form
        - 500 - if something else goes wrotn in the backend

    2.  Body - the submission object shown below

```typescript 
interface Submission {
    ID: uint;
}
```

### Submission Retrieval

Gets a submission by ID to send to the frontend

1. Enpoint

    The enpoint to upload submission is **GET** `/submissions/{id}` where ID
    is the submission ID as a uint

2. Request

    1. Headers - request needs no headers 

    2. Body - request has no body

3. Response

    1.  Status

        - 200 - if the submision is uploaded properly
        - 400 - if the submission is not sent in the right form
        - 500 - if something else goes wrotn in the backend

    2.  Body - the submission object shown below

```typescript
interface Submission {
    ID: uint;
    CreatedAt: DateTime;
    UpdatedAt: DateTime;
    DeletedAt: DateTime;
    name: string;
    license: string;
    files: File[];
    authors: GlobalUser[];
    reviewers: GlobalUser[];
    categories: string[];
    metaData: SubmissionData;
}

interface SubmissionData  {
    abstract: string;
    reviews: Comment[]
}

interface File {
    ID: uint;
    submissionId: uint;
    path: string;
    name: string;
}

interface GlobalUser {
    userId: string;
    fullName: string;
}
```

### Submission Download

Download a given submission as a zip archive. Available to any user

1. Endpoint

    The endpoint for downloads is **GET** `/submission/{id}/download` where id is submission ID as a uint

2. Request 

    1. Headers - n/a

    2. Body - n/a

3. Response

    1. Status

        - 200 - if the request succeeds and the zip is returned
        - 400 - if the request is malformatted (i.e. submission id is "a" or similar)
        - 404 - if the submission is not found
        - 500 - something has gone wrong in the server not relating to the request

    2. Body

        The response body is a base64 Standard Encoded version of the zip file with content type `application/zip`

## Approval

### Assigning Reviewers 

Editor assigns reviewers to a given submission

1. Endpoint

    The endpoint to upload a review is **POST** `/submission/{id}/assignreviewers` where ID is the submission ID as a uint

2. Request

    1. Headers 

    2. Body - the request body is shown below

```typescript
interface AssignReviewersBody {
    Reviewers: []string
}
```

3. Response

    1. Status

        - 200 - if the request fully succeeds
        - 400 - if the request is malformatted or one of the given userIDs does not have reviewer permissions
        - 401 - if the client does not have editor permissions
        - 409 - if the submission status has been finalised (i.e. approved/dissaproved)
        - 500 - unexpected error

    2. Body

### Review Upload

Uploads a review for a given reviewer of a submission

1. Endpoint

    The endpoint to upload a review is **POST** `/submission/{id}/review` where ID is the submission ID as a uint

2. Request

    1. Headers 

    2. Body - the request body is shown below

```typescript
interface UploadReviewBody {
	approved: bool;
	base64Value: string;
}
```

3. Response

    1. Status:

        - 200 - everything happened as expected
        - 400 - fields missing from request body, bad formatting, or duplicate review upload
        - 401 - user is not logged in or registered as a reviewer for the given submission
        - 500 - unexpected error

    2. Body:

### Update Submission Status

Changes the status of a submission to either approved or disapproved

1. Endpoint

    The endpoint to upload a review is **POST** `/submission/{id}/approve` where ID is the submission ID as a uint

2. Request

    1. Headers 

    2. Body - the request body is shown below

```typescript
interface UpdateSubmissionStatusBody {
    status: bool;
}
```

3. Response

    1. Status:

        - 200 - everything happened as expected
        - 400 - fields missing from request body, bad request body format, or submissionID not set properly in the URL
        - 401 - user is not logged in or registered as an editor (in practice this means the context is not set for the request)
        - 409 - not all reviewers have uploaded reviews yet. This is not allowed behaviour
        - 500 - any other miscellaneous errors

    2. Body:

## Comments

### User Comment Upload

Uploads a user comment/comment reply to a given file

1. Enpoint

    The enpoint to upload submission is **POST** `/file/{id}/newcomment` where ID
    is the submission ID as a uint

2. Request

    1. Headers - request needs no headers 

    2. Body - Typescript object shown below

```typescript
interface NewCommentPostBody {
	authorId: string
	parentId: *uint // optionally set for replies
	base64Value: string
}
```

3. Response

    1. Status

        - 200 : Comment Added Successfully
        - 400 : if the comment was not sent in the proper format
        - 500 : if something else goes wrong in the server

    2. Body - Typescript object shown below

```typescript
interface NewCommentResponse {
    id: uint
}
```

