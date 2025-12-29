# User service

A microservice for user authentication, registration, and management.

## Tech stack

- Go 1.25.5
- MongoDB
- Redis

## Development

Additional services are available for local development:

- `localhost:3000/dev/user-service` - serves [mongo-express](https://github.com/mongo-express/mongo-express)
- `localhost:3101` - direct connection to MongoDB

## Structure

- The purpose of this service is to handle role-based user authentication,
  user registration, log out with session termination, forcing good practices
  with occasional password resets, account recovery via magic link as well as
  access/refresh token issuing.

### The idea behind the data model and database choice

- We took time to do our research and have concluded on why is document-oriented
  database the best choice for this service. User entity requires storing nested
  data such as profile info, settings, roles physically close together in a single
  JSON/BSON document. This approach minimizes latency and maximizes reading performance,
  which is a core benefit provided by document-oriented DB's. By applying a unique
  secondary index on the email/username field, we ensure that the required user
  document can be located and fetched with high performance, making the login process
  efficient and reliable.

### User entity structure

```json
{
	"id": "UUID",
	"username": "string (unique)",
	"first_name": "string",
	"last_name": "string",
	"email": "string (unique)",
	"password": "string",
	"role": "MEMBER | ADMIN",
	"password_last_changed_at": "timestamp",
	"password_expires_at": "timestamp",
	"account_status": "ACTIVE | INACTIVE",
	"created_at": "timestamp",
	"updated_at": "timestamp",
    "last_expiry_notification_sent_at": "timestamp"
}
```
