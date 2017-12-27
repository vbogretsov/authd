# authd

OAuth 2.0 compliant authentication service

## API

### Create account

#### Request

POST /v1/requests/accounts HTTP/1.1
Content-Type: application/json; charset=UTF-8

{
    "email": "test-user-1@mail.com",
    "password": "123456"
}

#### Response

HTTP/1.1 201 Created
Content-Type: application/json; charset=UTF-8

{
    "user":
    {
        "id": "e832f45a-6faa-4eb4-8c40-7ad30ab93df8",
        "email": "test-user-1@mail.com",
        "active: false
    }
}

#### Errors

400 - if credentials do not meet criteria

### Confirm account

#### Request

GET /v1/requests/accounts/c7719bb1-f640-4cd7-9025-01b5ba1a7ad3 HTTP/1.1

#### Response

HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8

{
    "user":
    {
        "id": "e832f45a-6faa-4eb4-8c40-7ad30ab93df8",
        "email": "test-user-1@mail.com",
        "active: true
    }
    "token":
    {
        "access": "0a5c26d7-1c5c-4b37-a2bd-9d084152b33f"
        "refresh": "67381c43-51eb-43fb-be51-8ed63bb11050"
        "created": "2017-01-01 00:00:00"
        "expired": "2017-01-01 00:00:00"
    }
}

#### Errors

404 - if no confirmation with the id provided found

### Login

#### Request

POST /v1/tokens/ HTTP/1.1
Content-Type: application/json; charset=UTF-8

{
    "email": "test-user-1@mail.com",
    "password": "123456"
}

#### Response

HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8

{
    "access": "0a5c26d7-1c5c-4b37-a2bd-9d084152b33f"
    "refresh": "67381c43-51eb-43fb-be51-8ed63bb11050"
    "created": "2017-01-01 00:00:00"
    "expired": "2017-01-01 00:00:00"
}

#### Errors

401 - if credentials are invalid

### Request password reset

#### Request

POST /v1/requests/credentials/
Content-Type: application/json; charset=UTF-8

{
    "email": "test-user-1@mail.com"
}

#### Respnse

HTTP/1.1 204 No content

### Reset password

#### Request

POST /v1/requests/credentials/39802b61-9dd4-4ec5-b5c3-94f3d6c68ef0
Content-Type: application/json; charset=UTF-8

{
    "password": "123456"
}

#### Response

HTTP/1.1 204 No content

#### Errors

400 - if the password provided does not meet criteria
