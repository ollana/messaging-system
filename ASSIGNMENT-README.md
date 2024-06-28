# A messaging system backend in the cloud! - Cloud Computing Course Exercise 1

## Assignment
### The scenario:
The task you have is to build a messaging system, such as Telegram, WhatsApp, etc in the cloud.

You only need to provide a backend, no frontend required.

### Actions:
You need to provide endpoints that would allow the following actions:

- Register a new user (generating a new id)
- Send a message to a user (via its id):
  * Messages are sent from a user to a user
  * Check if the user is blocked from sending
- Allow a user to block another user from sending a message to them
- Creating a group
- Adding / removing users
- Sending messages to a group
- Users can check for their messages
  * Assume that users will check messages at least once a minute - consider scaling factors here.

### Notes:  
Pay attention to the fallacies of distributed computing, and the number of calls required to perform operations.
This task requires persistence. You are free to use whatever persistence technology you like, but your solution needs to also deploy that with your system.

**Required:** Include a discussion of the scaling effects on your system at 1000s of users, 10,000s or users and millions of users.

### Deliverables:
- Code that would handle the above-mentioned requirements.
- Discussion on the scaling of your system and how it reacts (both is load and in dollars) to more users & load.
- Include a script that would deploy the code to the cloud. Can be bash, cloud formation,
custom code, Pulumi, etc.
- Push the results to GitHub or similar service and provide a link to the code.
- Inclusion of access keys in the submission will automatically reduce 25% of the grade.
