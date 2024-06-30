# Messaging system - Cloud Computing Course Exercise 2

[Assignment details](ASSIGNMENT-README.md#assignment) | [Solution](#solution) | [Deployment steps](#deployment-steps)

## Solution

The service is implemented in [Golang](https://go.dev/), and [Pulumi](https://www.pulumi.com/) is used to deploy the service to AWS EC2.
I assume the service will not have much idle time as users checking for messages at least once a minute. 



### Database
AWS DynamoDB will be used as the database for the messaging system.
DynamoDB is a fully managed NoSQL database service that offers high performance, scalability, and low-latency consistency.
The considerations for cost should be weighed based on the expected workloads, and can be optimized further depending on traffic patterns.

##### The database was modeled according to service needs and access patterns.
- User table: 
  - userId (string) - HashKey
  - username (string)
  - blockedUsers (list of strings)
- Group table:
  - groupId (string) - HashKey
  - groupName (string)
  - users (list of strings)
- Message table:
  - recipientId (string) - HashKey 
  - timestamp (string) - SortKey
  - senderId (string)
  - message (string)
  
##### DB access for service calls:

Most frequent service call:
- check messages for a user
  - get user - 1 get call by HashKey
  - get messages - 1 get call by HashKey+SortKey

Frequent service calls:
  - send a message to a user
    - get both users - 2 get calls by HashKey
    - write message - 1 write call
  - send a message to a group 
    - get group and sender user - 2 get calls by HashKey
    - write message - 1 write call

Rare service calls:
  - create a new user/group 
    - write user/group - 1 write call  
  - add/remove user to group 
    - get user and group - 2 get calls by HashKey 
    - update group and user - 2 write call 
  - block/unblock a user
    - get both users - 2 get call by HashKey 
    - update user - 1 write call

##### Caching:
- Users and group cache - as this data will not change frequently, we can cache the data for a certain period of time, reducing the number of read calls to the database.
- Messages:
  - Unlike the users and group tables, messages are updated frequently. To avoid inconsistency, the cache needs to be centralized and should be updated on every write operation. This was not implemented as part of this submission but should be considered if the service is expecting high traffic to improve the performance and reduce cost.
  - Both direct and group messages are fetched for a user from the DB at once. For caching purposes, as messages from a group are likely to be read by all users of the group, we should cache the group messages to be reused by all users of the group. When fetched by a user, the service will filter and cache the group messages.  This will reduce the number of calls to the database.

With this design, we can optimize the number of calls to the database and reduce the cost of the service.
- for the most frequent service call, we only need 2 get calls to get the user and messages.

Let's focus on the most frequent service call: 
As we know every user will check for messages at least once a minute, we can assume that the service will have a high number of read requests which will only depend on the number of users. 

##### Scaling considerations:


### Assumptions:
- As it is not the main focus of the exercise, the service is not following the best security practices:
  - There is no authentication in the system.
  - The service is not using HTTPS.
  - Security groups settings are not optimized.
  - Authorization is basic and limited - for example user can only send message to groups they are part of.
  - There is no rate limiting.
  - And more...

#### Functionality assumptions:
*User*
- No authorization needed to create a user, nor blocking or unblocking users.
- User-name is not unique.

*Group*
- No authorization needed to create a group, nor adding or removing users to group.
- Group-name is not unique.
- If user is already in the group, adding again will return error.
- If user is not in the group, removing will return error.

*Block*
- Blocking already blocked user will return error.
- Block user will block from sending direct messages only. Blocked user can still send messages to groups they are part of.
- Blocked user will get forbidden error when trying to send a private message.
- User can block itself.
- 
*Message*
- User can send message to self, other users and groups it is part of.
- If user have no messages, the response will contain null message array.
- When getting messages, all user messages will be returned including private and group messages. It is up to the client to filter the messages based on the sender and recipient.
- Timestamp is unix representation of time in milliseconds and is optional - If no timestamp is provided, all messages will be returned.
- If timestamp is provided, messages after the timestamp will be returned. We assume the client will always request for messages after the last timestamp received or last timestamp requested.
- User will get self sent messages as well, including group messages they sent.


### APIs:

- Create a New User 
  ```
  POST /v1/users/create
  Request:  { "username": "string" }
  Response: { "userId": "string", "username": "string" }
  ```
  
- Block a User
  ```
  POST /v1/users/:userId?op=block
  Request:  { "blockedUserId": "string" }
  ```
  
- Unblock a User
  ```
    POST /v1/users/:userId?op=unblock
    Request:  { "blockedUserId": "string" }
    ```

- Create a Group
    ```
    POST /v1/groups/create
    Request:  { "groupName": "string" }
    Response: { "groupId": "string", "groupName": "string" }
    ```
  
- Add User to Group
    ```
    POST /v1/groups/:groupId?op=add
    Request:  { "userId": "string" }
    ```
- Remove User from Group  
    ```
    POST /v1/groups/:groupId?op=remove
    Request:  { "userId": "string" }
    ```

- Send a Message to a User
  ```
  POST /v1/messages/send?type=private
  Request:  { "senderId": "string", "receiverId": "string", "message": "string" }
  ```
  
- Send a Message to a Group
    ```
    POST /v1/messages/send?type=group
    Request:  { "senderId": "string", "groupId": "string", "message": "string" }
    ```
  
- Check All Messages for a User
    ```
    GET /v1/messages/:userId
    Response: { "messages": [ { "senderId": "string", "message": "string", "recipientId": "string", timestamp": "string" } ] }
    ```
- Check Messages for a User from last timestamp
    ```
    GET /v1/messages/:userId?timestamp=123456789
    Response: { "messages": [ { "senderId": "string", "message": "string", "recipientId": "string", timestamp": "string" } ] }
    ```

## Deployment steps

Using makefile
#### Prerequisites
- [make](https://www.incredibuild.com/integrations/gnu-make)
- [golang 1.22+](https://go.dev/doc/install)
- [pulumi 3.0.1+](https://www.pulumi.com/docs/install/)
- aws credentials setup to a profile named `pulumi`

#### Steps
1. login to pulumi
``` bash
pulumi login
```
2. insert your aws account id as an environment variable
```
export AWS_ACCOUNT_ID=<aws-account-id>
```
3. deploy the service
``` bash 
make deploy
```
4. destroy the service
``` bash
make destroy
```

