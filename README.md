# Messaging system - Cloud Computing Course Exercise 2

[Assignment details](ASSIGNMENT-README.md#assignment) | [Solution](#solution) | [Deployment steps](#deployment-steps)

## Solution


The service is implemented in [Golang](https://go.dev/), and [Pulumi](https://www.pulumi.com/) is used to deploy the service to AWS EC2.
I assume the service will not have much idle time as users checking for messages at least once a minute. 

AWS DynamoDB will be used as the database for a messaging system. 
DynamoDB is a fully managed NoSQL database service that offers high performance, scalability, and low-latency consistency.
The considerations for cost should be weighed based on the expected workloads, and can be optimized further depending on traffic patterns.

### Additional assumptions:

- Any op on non-existing group or user will return error.
- There is no authentication in the system.
- Authorization is basic - user can only send message to groups they are part of.

*User*
- No authorization needed to create a user, nor blocking or unblocking users.
- User-name is not unique.

*Block*
- Blocking already blocked user will return error.
- Block user will block from sending direct messages only. Blocked user can still send messages to groups they are part of.
- Blocked user will get forbidden error when trying to send a private message.
- User can block itself.

*Group*
- No authorization needed to create a group, nor adding or removing users to group.
- Group-name is not unique.
- If user is already in the group, adding again will return error.
- If user is not in the group, removing will return error.

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

