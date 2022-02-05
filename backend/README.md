
# Table of Contents

1.  [Response Format](#org22af552)
2.  [API](#orgf6ba78d)
    1.  [Journal](#org427c0d4)
        1.  [Login](#org1f3d312)
    2.  [Authentication](#orgcbad998)
        1.  [Login](#org2ae3186)



<a id="org22af552"></a>

# Response Format

Response bodies are always given in this below format (with the exception of journal endpoints) -

    interface Response {
    	message: string // Response's message - gives details on response status.
    	error: boolean  // Whether the request errored or not.
    	content: any    // The response's body.
    }


<a id="orgf6ba78d"></a>

# API

This section describes main enpoints. There are 4 main entrypoints in the api,
starting with `cs3099user11.host.cs.st-andrews.ac.uk/api/<version>`:

-   `/journal/`: Supergroup compliant endpoints,
-   `/auth/`: Authentication requests - client login, signup.
-   `/user/`: User retrieval requests.


<a id="org427c0d4"></a>

## Journal


<a id="org1f3d312"></a>

### Login

Supergroup compliant journal login.

1.  Endpoint

    This is accessed via the **GET** `/journal/login` endpoint.

2.  Request

    1.  Headers
    
        -   X-FOREIGNJOURNAL-SECURITY-TOKEN: Journal&rsquo;s secret key. Needed for all journal queries. **MUST NOT BE SHARED WITH ANY CLIENT.**
    
    2.  Body
    
            {
            	"email": "string",
            	"password": "string"
            }

3.  Response

    1.  Status
    
        -   200 - User logged in, response OK.
        -   404 - User was not found.
        -   401 - Access Unauthorized. See Headers.
    
    2.  Content
    
            interface Content {
            	userId: string   // ID of the logged in user. Empty if errored
            }


<a id="orgcbad998"></a>

## Authentication


<a id="org2ae3186"></a>

### Login

Client-only login endpoint.

1.  Endpoint

    This is accessed via the **GET** `/auth/login` endpoint.

2.  Request

    1.  Headers
    
        No headers required.
    
    2.  Body
    
            interface Request {
            	client_id: string 		// ID used to find client - here, it is an email.
            	client_secret: string   // Client's secret for authentication - here, password.
            }

3.  Response

    1.  Status
    
        -   200 - User logged in, response OK.
        -   404 - User was not found.
        -   401 - Access Unauthorized. See Headers.
    
    2.  Content
    
            interface Content {
            	token: string   			// Token used for restriced user access later on.
            	refresh_token: string		// Refresh token - used to make a new token after expiry.
            	redirect_uri: string 		// Address to token refresh request.
            }

