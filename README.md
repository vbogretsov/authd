# authd

Simple authentication service with email confirmation and password reset

## API

### Create account

#### Request

```http
POST /signup HTTP/1.1
Content-Type: application/json; charset=UTF-8

{
    "email": "test-user-1@mail.com",
    "password": "123456"
}
```

#### Response

```http
HTTP/1.1 201 Created
Content-Type: application/json; charset=UTF-8

{
    "info": "confirmation email has been sent to your email"
}
```

#### Errors

 * 400 - if credentials do not meet criteria
 * 409 - if email already exists

### Confirm account

#### Request


```http
POST /singup/4262c357-e8a7-4714-9327-f9e3610e0fc4 HTTP/1.1
```

#### Response

```
HTTP 200 OK
Content-Type: application/json; charset=UTF-8

{
    "info": "account has been confirmed"
}
```

#### Errors

 * 404 - if no confirmation with the id provided was found

### Login

#### Request

```http
POST /signin HTTP/1.1
Content-Type: application/json; charset=UTF-8

{
    "email": "test-user-1@mail.com",
    "password": "123456"
}
```

#### Response

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8

{
    "token":
    {
        "created": "2017-01-01 00:00:00",
        "expires": "2017-01-02 00:00:00",
        "accessToken": "???",
        "refreshToken": "???"
    }
}
```

#### Errors

 * 401 - if credentials are invalid or email is not confirmed

### Logout

#### Request

```http
POST /signout HTTP/1.1
Authorization: Bearer ???
```

#### Response

```http
HTTP/1.1 204 No content
```

### Refresh token

#### Request

```http
POST /refresh HTTP/1.1
Content-Type: application/json; charset=UTF-8

{
    "refreshToken": "???"
}

```

#### Response

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8

{
    "token":
    {
        "created": "2017-01-01 00:00:00",
        "expires": "2017-01-02 00:00:00",
        "accessToken": "???",
        "refreshToken": "???"
    }
}
```

### Request password reset

#### Request

```http
POST /pwreset
Content-Type: application/json; charset=UTF-8

{
    "email": "test-user-1@mail.com",
}
```

#### Respnse

```http
HTTP/1.1 200 OK

{
    "info": "password rese email has been sent to your email"
}
```

### Reset password

#### Request

```http
POST /pwreset/39802b61-9dd4-4ec5-b5c3-94f3d6c68ef0
Content-Type: application/json; charset=UTF-8

{
    "password": "123456"
}
```

#### Response

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8

{
    "info": "password has been updated"
}
```

#### Errors

 * 400 - if the password provided does not meet criteria
 * 404 - if password reset request with the id provided was not found


### Login via third party providers

#### Google

TODO

#### GitHub

TODO

#### Facebook

TODO
