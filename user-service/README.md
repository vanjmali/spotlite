# User service

- The purpose of this service is to handle role-based user authentication, 
user registration, log out with session termination, forcing good practices
with occasional password resets, account recovery via magic link as well as
access/refresh token issuing.

## The idea behind the data model and database choice

- We took time to do our research and have concluded on why is document-oriented 
database the best choice for this service. User entity requires storing nested 
data such as profile info, settings, roles physically close together in a single
JSON/BSON document. This approach minimizes latency and maximizes reading performance, 
which is a core benefit provided by document-oriented DB's. By applying a unique
secondary index on the email/username field, we ensure that the required user
document can be located and fetched with high performance, making the login process
efficient and reliable.

## User entity structure

{
  "id": "UUID",
  "username": "string (unique)",
  "firstName": "string",
  "lastName": "string",
  "email": "string (unique)",
  "password": "string",
  "role": "MEMBER | ADMIN",
  "passwordLastChanged": "timestamp",
  "accountStatus": "ACTIVE | INACTIVE",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
