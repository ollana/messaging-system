# Messaging system - Cloud Computing Course Exercise 2

[Assignment details](ASSIGNMENT-README.md#assignment) | [Solution](#solution) | [Discussion of Scaling Effects](#discussion-of-scaling-effects) |  [Deployment steps](#deployment-steps)

## Solution

The service is implemented in [Golang](https://go.dev/), and [Pulumi](https://www.pulumi.com/) is used to deploy the service to AWS EC2.
DynamoDB is used as the database for the messaging system.
I assume the service will not have much idle time as users checking for messages at least once a minute.

###  Implementation assumptions:
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
- User will get messages from groups they are currently part of. If user is removed from group, they will not get any messages from that group, even if the user was part of the group when it was sent.

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
  - get messages - 1 get call by HashKey+SortKey + x get calls by HashKey+SortKey for group messages depending on the number of groups the user is part of

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
- Users and group cache - as this data will not change frequently, we will cache the data for a certain period of time, reducing the number of read calls to the database.
- Messages:
  - User messages cache will not be efficient as we don't expect the same user requesting the same messages more than once.
  - Group messages are cached so all users of a group can read the messages from the cache. This will reduce the number of calls to the database.
  - The cache updates on every write operation. As we are starting with only one instance, we can use the in-memory cache and keep it up to date.
  - Cache evicts based on last access time, this way we keep the most popular groups in the cache. 
  - Only recent messages per group is stored - any messages older than one minute are evicted. As usually users will check for messages at least once a minute, it will be rare to check for messages older than a minute.
  - For scaling, to avoid inconsistency, the cache needs to be centralized and should be updated on every write operation. This was not implemented as part of this submission but should be considered if the service is expecting high traffic to improve the performance and reduce cost. 

## Discussion of Scaling Effects:
in the following section, we will discuss how user scaling for each scenario will affect the system load and cost.
Later we will break down the estimated cost for each scenario, with modification being made for optimization.

##### Assumptions for cost estimation:
###### DynamoDB
Assuming the pay-per-request model:

RCU: $0.00013 per RCU-hour
WCU: $0.00065 per WCU-hour

###### ECS/Fargate
vCPU: $0.04048 per hour
GB Memory: $0.004445 per hour 

###### Amazon ElastiCache Redis
`cache.t2.micro` instance: $0.017 per hour


######  Common operations read/write usage:
- Checking Messages
    - Every user checks messages about once a minute.
    - Assume an average user checks 5 individual messages + 2 group messages per check.
    - Read operations: 7 reads per minute per user.
    - For smaller scales, simple in-memory caching is used. Assume caching reduces DynamoDB read operations by 30%.
    - At larger scales, centralized and scalable cache mechanisms service like Amazon ElastiCache with Redis will be used. Assume caching reduces DynamoDB read operations by 70%.
- Sending Messages
    - Assume each user sends 10 messages a day.
    - Write operations: 10 writes per day per user.
- Other operation are negligible compared to the above operations.

### Scaling to Thousands of Users (1,000 - 10,000 Users)
##### System Load
###### Database Load:

- Read Operations: For typical user interactions (sending messages, checking messages), thousands of users should not impose a significant load on DynamoDB given its ability to handle large numbers of read/write operations per second. With caching in place, read operations on user and group data are optimized.
- Write Operations: The distribution of messages across users and groups ensures that write operations are spread out, mitigating any single point of contention.
Cache Efficiency: In-memory caching for user and group data significantly reduces database read operations, maintaining low latency.
###### Compute Load:

A single Fargate service instance with ECS should suffice for thousands of users, handling HTTP requests efficiently.
CPU and memory usage is relatively low, given typical user activity rates.
### Cost
- DynamoDB: Costs are relatively low because of the pay-per-request billing mode; the usage patterns do not induce high costs at this scale.
- ECS/Fargate: Costs are manageable with a single instance for compute needs.
- Cache: Minimal additional costs for in-memory caching within the ECS container.

#### ECS/Fargate cost and optimization:
##### Adjusted Configuration:
 - vCPU: 0.5 (doubling to handle more concurrent requests)
 - Memory: 1 GB (doubling to handle more concurrent users and operations)

 - vCPU: 0.5 vCPU * $0.04048 * 24 hours * 30 days = $14.57/month
 - Memory: 1 GB * $0.004445 * 24 hours * 30 days = $3.20/month
###### Monthly Cost: $17.77

#### DynamoDB cost and optimization:
- Checking Messages - Total reads per second: (7 reads/minute * 10,000 users) / 60 seconds * 0.7 = ~816 reads/second. 
- Sending Messages - Total writes per second: 100,000 writes/day / 86400 seconds/day = 1.16 writes/second.

- RCUs: Round to 1,000 for burst capacity
- WCUs: Round up to 5 for burst capacity.

- DynamoDB RCUs: 1,000 RCUs * $0.00013/hour * 24 hours/day * 30 days = $93.60/month.
- DynamoDB WCUs: 5 WCUs * $0.00065/hour * 24 hours/day * 30 days = $2.34/month.
- ###### Monthly Cost: $95.54

### Total Monthly Cost: $113.31 

### Scaling to Tens of Thousands of Users (10,000 - 100,000 Users)
##### System Load
###### Database Load:

- Read Operations: Scaling to tens of thousands of users might require more sophisticated caching mechanisms to maintain low latency; consider moving to distributed caching solutions like Amazon ElastiCache for centralization.
- Write Operations: Increased write operations due to higher user engagement necessitate optimized write patterns or DynamoDB auto-scaling capabilities.
###### Compute Load:

The ECS cluster should be scaled horizontally by increasing the number of Fargate instances to handle increased HTTP request loads.
Implementing load balancing effectively distributes traffic, maintaining high availability and performance.
#### Cost
- DynamoDB: Likely higher costs due to increased read/write throughput but still manageable with optimized access patterns.
- ECS/Fargate: Costs increase proportionally with the number of instances required to maintain performance.
- ElastiCache: Additional costs for implementing distributed caching but significantly improves read performance and decreases DynamoDB costs.

#### ECS/Fargate cost and optimization:
##### Adjusted Configuration:
- vCPU: 1
- Memory: 2 GB 
- Instances: 3 - to handle increased traffic and maintain high availability.

- vCPU: 1 vCPU * $0.04048 * 24 hours * 30 days * 3 instances = $87.43/month
- Memory: 2 GB * $0.004445 * 24 hours * 30 days * 3 instances = $19.20/month
###### Monthly Cost:  $106.63

#### DynamoDB cost and optimization:
- Checking Messages - Total reads per second: (7 reads/minute * 100,000 users) / 60 seconds * 0.3 = ~3,500 reads/second.
- Sending Messages - Total writes per second: 100,000,000 writes/day / 86400 seconds/day = 12 writes/second.

- RCUs: Round to 3,600 for burst capacity
- WCUs: Round up to 15 for burst capacity.

- DynamoDB RCUs: 3,600 RCUs * $0.00013/hour * 24 hours/day * 30 days = $336.96/month.
- DynamoDB WCUs: 15 WCUs * $0.00065/hour * 24 hours/day * 30 days = $7.02/month.
###### Monthly Cost: $343.98

#### Amazon ElastiCache Redis:
Number of Redis Nodes: 3 nodes (to distribute the load and maintain high availability).

- $0.017 per hour * 24 hours/day * 30 days * 3 = $36.72/month.
###### Monthly Cost: $36.72

### Total Monthly Cost: $487.33

### Scaling to Millions of Users
##### System Load
##### Database Load:

- Read Operations: With millions of users, the system must handle a massive volume of reads. Optimizations include:
Sharding data in DynamoDB to distribute load.
Extensive use of caching with advanced strategies like cache invalidation and partitioning.
- Write Operations: Writes must be carefully managed. Batch operations, write leveling, and optimized indexing can help manage the load on DynamoDB.
######  Compute Load:

Horizontal scaling becomes critical. Multiple ECS Fargate instances across different AZs (Availability Zones) ensure reliability and load distribution.
Advanced load balancers (e.g., ALB or ELB) further mitigate single points of failure and distribute incoming traffic efficiently.
#### Cost
- DynamoDB: Significant costs due to high read/write throughput. Consider Reserved Capacity or Provisioned Throughput billing modes if usage patterns are predictable.
- ECS/Fargate: Increased costs due to multiple instances but essential for ensuring service availability and reliability.
- ElastiCache: Higher costs for a large-scale distributed cache, accounting for a considerable part of the overall cost but necessary for maintaining performance.
- Networking Costs: Elevated due to data transfer between distributed services and users.

#### ECS/Fargate cost and optimization:
##### Adjusted Configuration:
- vCPU: 2
- Memory: 4 GB
- Instances: 12 - to handle increased traffic and maintain high availability.

- vCPU: 2 vCPU * $0.04048 * 24 hours * 30 days * 12 instances = $699.49/month
- Memory: 4 GB * $0.004445 * 24 hours * 30 days * 12 instances = $153.61/month
###### Monthly Cost:  $853.10

#### DynamoDB cost and optimization:
- Checking Messages - Total reads per second: (7 reads/minute * 1,000,000 users) / 60 seconds * 0.3 = ~35,000 reads/second.
- Sending Messages - Total writes per second: 10,00,000,000 writes/day / 86400 seconds/day = 116 writes/second.

- RCUs: Round to 36,000 for burst capacity
- WCUs: Round up to 120 for burst capacity.

- DynamoDB RCUs: 36,000 RCUs * $0.00013/hour * 24 hours/day * 30 days = $3369.6/month.
- DynamoDB WCUs: 120 WCUs * $0.00065/hour * 24 hours/day * 30 days = $56.16/month.
###### Monthly Cost: $3425.76

#### Amazon ElastiCache Redis:
Number of Redis Nodes: 10 nodes (to handle the significantly higher load).

- $0.017 per hour * 24 hours/day * 30 days * 10 = $122.40/month.
###### Monthly Cost: $122.40

### Total Monthly Cost: $4401.26

Cost summery for each scenario:
- Thousands of Users: $113.31
- Tens of Thousands of Users: $487.33
- Millions of Users: $4401.26


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

