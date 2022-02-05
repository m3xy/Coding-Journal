
# Table of Contents

1.  [Response Format](#org53d091f)
2.  [API](#orgb09e5ff)
    1.  [Journal](#org6978210)
        1.  [Login](#org6a5f21c)
    2.  [Authentication](#orgc9014e9)
        1.  [Login](#orgc990c37)
        2.  [Register](#orgb295fc3)
    3.  [User](#org7ea7e4b)
        1.  [Profile Query](#org1b33072)



<a id="org53d091f"></a>

# Response Format

Response bodies are always given in this below format (with the exception of journal endpoints) -

    interface Response {
    	message: string // Response's message - gives details on response status.
    	error: boolean  // Whether the request errored or not.
    	content: any    // The response's body.
    }


<a id="orgb09e5ff"></a>

# API

This section describes main enpoints. There are 4 main entrypoints in the api,
starting with `cs3099user11.host.cs.st-andrews.ac.uk/api/<version>`:

-   `/journal/`: Supergroup compliant endpoints,
-   `/auth/`: Authentication requests - client login, signup.
-   `/user/`: User retrieval requests.


<a id="org6978210"></a>

## Journal


<a id="org6a5f21c"></a>

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


<a id="orgc9014e9"></a>

## Authentication


<a id="orgc990c37"></a>

### Login

Client-only login endpoint.

1.  Endpoint

    This is accessed via the **GET** `/auth/login` endpoint.

2.  Request

    1.  Headers
    
        No headers required.
    
    2.  Body
    
            interface Body {
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


<a id="orgb295fc3"></a>

### Register

Client registration endpoint.

1.  Endpoint

    The endpoint is accessible as **GET** `/auth/register`

2.  Request

    1.  Headers
    
        No required headers.
    
    2.  Body
    
            interface Content {
            	// Required
            	Email: string  	// Client ID
            	Password: string    // Client Secret
            	FirstName: string
            	LastName: string
            
            	// Optional
            	PhoneNumber?: string
            	Organization?: string
            }

3.  Response

    1.  Status
    
        -   200 - User successfully registered, response OK.
        -   405 - Bad request, form given is invalid.
    
    2.  Body
    
        This function has no response content.


<a id="org7ea7e4b"></a>

## User


<a id="org1b33072"></a>

### Profile Query

User profile information query.

1.  Endpoint

    The endpoint for the query is `/user/{ID}`, where ID stands for the user&rsquo;s UUID in the server.

2.  Request

    1.  Headers
    
        No header or authentication is required for this query.

3.  Response

    1.  Status
    
        -   200 - Query successful, requested information contained in content.
        -   404 - User not found, content empty.
    
    2.  Body
    
        The \`Content\` value in the response is of type below
        
            interface Content {
            	UserID: string
            	FullName: string
            	Profile: Profile
            }
            
            interface Profile {
            	Email: string
            	FirstName: string
            	LastName: string
            	PhoneNumber: string
            	Organization: string
            	CreatedAt: DateTime // Format -
            }

