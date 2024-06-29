# Messaging system - Cloud Computing Course Exercise 2

[Assignment details](ASSIGNMENT-README.md#assignment) | [Solution](#solution) | [Deployment steps](#deployment-steps)

## Solution


The service is implemented in [Golang](https://go.dev/), and [Pulumi](https://www.pulumi.com/) is used to deploy the service to AWS EC2.
I assume the service will not have much idle time as users checking for messages at least once a minute. 

AWS DynamoDB will be used as the database for a messaging system. 
DynamoDB is a fully managed NoSQL database service that offers high performance, scalability, and low-latency consistency.
The considerations for cost should be weighed based on the expected workloads, and can be optimized further depending on traffic patterns.

### Additional assumptions:
- User-name and group-name are not unique
- Any op on non-existing group or user will return error
- There is no authentication in the system, basic authorization are based on the user id.

*Block*
- Blocking already blocked user will return error.
- Block user will block from sending direct messages only. Blocked user can still send messages to groups they are part of.
- Blocked user will get forbidden error when trying to send a private message.

*Group*
- Anyone can create/add to/remove from a group
- If user is already in the group, adding again will return error
- If user is not in the group, removing will return error
- User can send message to a group they are part of.

*Get Message*
- If user have no messages, the response will be empty message array
- Timestamp is unix representation of time in milliseconds and is optional - If no timestamp is provided, all messages will be returned.
- If timestamp is provided, messages after the timestamp will be returned. We assume the client will always request for messages after the last timestamp received or last timestamp requested.
- When getting messages, all user messages will be returned including group messages. It is up to the client to filter the messages based on the sender and recipient.
- User will get self sent messages as well, including group messages they sent.


### APIs:

- Register a New User 
  ```
  POST /v1/users/register
  Request:  { "Username": "string" }
  Response: { "UserId": "string", "Username": "string" }
  ```

- Send a Message to a User
  ```
  POST /v1/messages/send?type=private
  Request:  { "SenderId": "string", "ReceiverId": "string", "Message": "string" }
  ```
  
- Block a User
  ```
  POST /v1/users/:userId?op=block
  Request:  { "BlockedUserId": "string" }
  ```
  
- Unblock a User
  ```
    POST /v1/users/:userId?op=unblock
    Request:  { "BlockedUserId": "string" }
    ```

- Create a Group
    ```
    POST /v1/groups/create
    Request:  { "GroupName": "string" }
    Response: { "GroupId": "string", "GroupName": "string" }
    ```
  
- Add User to Group
    ```
    POST /v1/groups/:groupId?op=add
    Request:  { "UserId": "string" }
    ```
- Remove User from Group  
    ```
    POST /v1/groups/:groupId?op=remove
    Request:  { "UserId": "string" }
    ```
  
- Send a Message to a Group
    ```
    POST /v1/messages/send?type=group
    Request:  { "SenderId": "string", "GroupId": "string", "Message": "string" }
    ```
  
- Check All Messages for a User
    ```
    GET /v1/messages/:userId
    Response: { "Messages": [ { "SenderId": "string", "Message": "string", "RecipientId": "string", Timestamp": "string" } ] }
    ```
- Check Messages for a User from last timestamp
    ```
    GET /v1/messages/:userId?timestamp=123456789
    Response: { "Messages": [ { "SenderId": "string", "Message": "string", "RecipientId": "string", Timestamp": "string" } ] }
    ```

## Deployment steps

