# Messaging system - Cloud Computing Course Exercise 2

[Assignment details](ASSIGNMENT-README.md#assignment) | [Solution](#solution) | [Deployment steps](#deployment-steps)

## Solution


The service is implemented in [Golang](https://go.dev/), and [Pulumi](https://www.pulumi.com/) is used to deploy the service to AWS EC2.
I assume the service will not have much idle time as users checking for messages at least once a minute. 

AWS DynamoDB will be used as the database for a messaging system. 
DynamoDB is a fully managed NoSQL database service that offers high performance, scalability, and low-latency consistency.
The considerations for cost should be weighed based on the expected workloads, and can be optimized further depending on traffic patterns.

Additional assumptions:
- user-name and group-name are not unique
- blocking already blocked user will return error
- any op on non-existing group or user will return error
- there is no authentication or authorization in the system for simplicity
- anyone can create/add to/remove from a group
- If user is already in the group, adding again will return error
- If user is not in the group, removing will return error


### APIs:

- Register a New User 
  ```
  POST /v1/users/register
  Request:  { "username": "string" }
  Response: { "userId": "string", "username": "string" }
  ```

- Send a Message to a User
  ```
  POST /v1/messages/private
  Request:  { "senderId": "string", "receiverId": "string", "message": "string" }
  ```
  
- Block a User
  ```
  POST /v1/users/:userId/block
  Request:  { "blockedUserId": "string" }
  ```

- Create a Group
    ```
    POST /v1/groups/create
    Request:  { "groupName": "string", "creatorId": "string" }
    Response: { "groupId": "string", "groupName": "string" }
    ```
  
- Add User to Group
    ```
    POST /v1/groups/:groupId/add
    Request:  { "userId": "string" }
    ```
- Remove User from Group  
    ```
    POST /v1/groups/:groupId/remove
    Request:  { "userId": "string" }
    ```
  
- Send a Message to a Group
    ```
    POST /v1/messages/group
    Request:  { "senderId": "string", "groupId": "string", "message": "string" }
    ```
  
- Check Messages for a User
    ```
    GET /v1/messages/:userId
    Response: { "messages": [ { "senderId": "string", "message": "string" } ] }
    ```


## Deployment steps

